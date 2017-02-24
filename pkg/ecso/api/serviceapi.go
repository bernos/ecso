package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceAPI interface {
	DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error)
	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error)
	GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error)
}

type ServiceDescription struct {
	Name                     string
	URL                      string
	CloudFormationConsoleURL string
	CloudWatchLogsConsoleURL string
	ECSConsoleURL            string
	CloudFormationOutputs    map[string]string
}

// New creates a new API
func NewServiceAPI(cfg *ecso.Config) ServiceAPI {
	return &serviceAPI{cfg}
}

type serviceAPI struct {
	cfg *ecso.Config
}

func (api *serviceAPI) GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error) {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	var (
		log    = api.cfg.Logger()
		cfn    = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
		ecsAPI = reg.ECSAPI()
	)

	outputs, err := cfn.GetStackOutputs(s.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	if serviceName, ok := outputs["Service"]; ok {
		resp, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
			Cluster: aws.String(env.GetClusterName()),
			Services: []*string{
				aws.String(serviceName),
			},
		})

		if err != nil {
			return nil, err
		}

		if len(resp.Services) > 1 {
			return nil, fmt.Errorf("More than one service named '%s' was found", serviceName)
		}

		return resp.Services[0], nil
	}

	return nil, nil
}

func (api *serviceAPI) DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error) {
	var (
		log = api.cfg.Logger()
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())

	envOutputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	serviceOutputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	desc := &ServiceDescription{
		Name:                     service.Name,
		ECSConsoleURL:            util.ServiceConsoleURL(serviceOutputs["Service"], env.GetClusterName(), env.Region),
		CloudFormationConsoleURL: util.CloudFormationConsoleURL(service.GetCloudFormationStackName(env), env.Region),
		CloudWatchLogsConsoleURL: util.CloudWatchLogsConsoleURL(serviceOutputs["CloudWatchLogsGroup"], env.Region),
		CloudFormationOutputs:    make(map[string]string),
	}

	if service.Route != "" {
		desc.URL = fmt.Sprintf("http://%s%s", envOutputs["RecordSet"], service.Route)
	}

	for k, v := range serviceOutputs {
		desc.CloudFormationOutputs[k] = v
	}

	return desc, nil
}

func (api *serviceAPI) ServiceDown(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	log := api.cfg.Logger()
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	if err := api.deleteServiceStack(reg, env, service); err != nil {
		return err
	}

	log.Printf("\n")

	if err := api.clearServiceDNSRecords(reg, env, service); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cwLogsAPI := reg.CloudWatchLogsAPI()

	resp, err := cwLogsAPI.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(s.GetCloudWatchLogGroupName(env)),
	})

	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}

func (api *serviceAPI) ServiceUp(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	// set env vars so that they are available when converting the docker
	// compose file to a task definition
	if err := api.setEnv(project, env, service); err != nil {
		return err
	}

	// register task
	taskDefinition, err := api.registerECSTaskDefinition(reg, project, env, service)

	if err != nil {
		return err
	}

	// deploy the service cfn stack
	if err := api.deployServiceStack(reg, project, env, service, taskDefinition); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) setEnv(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	if err := util.AnyError(
		os.Setenv("ECSO_ENVIRONMENT", env.Name),
		os.Setenv("ECSO_AWS_REGION", env.Region),
		os.Setenv("ECSO_CLUSTER_NAME", env.GetClusterName())); err != nil {
		return err
	}

	// set any env vars from the service configuration for the current environment
	for k, v := range service.Environments[env.Name].Env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (api *serviceAPI) deployServiceStack(reg *ecso.AWSClientRegistry, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition) error {
	var (
		cfg       = api.cfg
		log       = cfg.Logger()
		stackName = service.GetCloudFormationStackName(env)
		prefix    = service.GetCloudFormationBucketPrefix(env)
		template  = service.GetCloudFormationTemplateFile()
		cfn       = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
	)

	params, err := getServiceStackParameters(cfn, project, env, service, taskDefinition)

	if err != nil {
		return err
	}

	tags := getServiceStackTags(project, env, service)

	log.Infof("Deploying service cloudformation stack '%s'...", stackName)

	result, err := cfn.PackageAndDeploy(
		stackName,
		template,
		prefix,
		tags,
		params,
		false)

	if err != nil {
		return err
	}

	if !result.DidRequireUpdating {
		log.Printf("  No updates were required to Cloud Formation stack '%s'\n", result.StackID)
	}

	return nil
}

func getServiceStackParameters(cfn helpers.CloudFormationHelper, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition) (map[string]string, error) {

	outputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"Cluster":        outputs["Cluster"],
		"AlertsTopic":    outputs["AlertsTopic"],
		"DesiredCount":   fmt.Sprintf("%d", service.DesiredCount),
		"TaskDefinition": *taskDefinition.TaskDefinitionArn,
	}

	if len(service.Route) > 0 {
		params["VPC"] = outputs["VPC"]
		params["Listener"] = outputs["Listener"]
		params["Path"] = service.Route
		params["Port"] = fmt.Sprintf("%d", service.Port)
		params["RoutePriority"] = strconv.Itoa(service.RoutePriority)
	}

	for k, v := range service.Environments[env.Name].CloudFormationParameters {
		params[k] = v
	}

	return params, nil
}

func getServiceStackTags(project *ecso.Project, env *ecso.Environment, service *ecso.Service) map[string]string {
	tags := map[string]string{
		"project":     project.Name,
		"environment": env.Name,
	}

	for k, v := range service.Tags {
		tags[k] = v
	}

	return tags
}

func (api *serviceAPI) registerECSTaskDefinition(reg *ecso.AWSClientRegistry, project *ecso.Project, env *ecso.Environment, service *ecso.Service) (*ecs.TaskDefinition, error) {
	var (
		cfg       = api.cfg
		log       = cfg.Logger()
		taskName  = service.GetECSTaskDefinitionName(env)
		ecsClient = reg.ECSAPI()
	)

	// TODO: fully qualify the path to the service compose file
	// taskDefinition, err := ConvertToTaskDefinition(taskName, service.ComposeFile)
	log.Infof("Converting '%s' to task definition...", service.ComposeFile)

	taskDefinition, err := service.GetECSTaskDefinition(env)

	log.Printf("\n")
	log.Infof("Registering ECS task definition '%s'...", taskName)

	if err != nil {
		return nil, err
	}

	for _, container := range taskDefinition.ContainerDefinitions {
		container.SetLogConfiguration(&ecs.LogConfiguration{
			LogDriver: aws.String(ecs.LogDriverAwslogs),
			Options: map[string]*string{
				"awslogs-region": aws.String(env.Region),
				"awslogs-group":  aws.String(service.GetCloudWatchLogGroupName(env)),
			},
		})

		// TODO probably don't automatically add service discovery env config
		//      if people want to use service discover they can just add the
		//      required env vars to the helpers docker-compose file. This is
		//      less magic and more flexible
		for _, p := range container.PortMappings {
			container.Environment = append(container.Environment, &ecs.KeyValuePair{
				Name:  aws.String(fmt.Sprintf("SERVICE_%d_NAME", *p.ContainerPort)),
				Value: aws.String(fmt.Sprintf("%s.%s", service.Name, env.GetClusterName())),
			})
		}
	}

	resp, err := ecsClient.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: taskDefinition.ContainerDefinitions,
		Family:               taskDefinition.Family,
		NetworkMode:          taskDefinition.NetworkMode,
		PlacementConstraints: taskDefinition.PlacementConstraints,
		TaskRoleArn:          taskDefinition.TaskRoleArn,
		Volumes:              taskDefinition.Volumes,
	})

	if err != nil {
		return nil, err
	}

	log.Printf(
		"  Registered ECS task definition %s:%d\n\n",
		*resp.TaskDefinition.Family,
		*resp.TaskDefinition.Revision)

	return resp.TaskDefinition, nil
}
func (api *serviceAPI) clearServiceDNSRecords(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log       = api.cfg.Logger()
		r53Helper = helpers.NewRoute53Helper(reg.Route53API(), log.Child())
		dnsName   = fmt.Sprintf("%s.%s.", service.Name, env.CloudFormationParameters["DNSZone"])
	)

	log.Infof("Deleting any service SRV DNS records for %s...", dnsName)

	if err := r53Helper.DeleteResourceRecordSetsByName(dnsName, env.CloudFormationParameters["DNSZone"], "Deleted by ecso service down"); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) deleteServiceStack(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log   = api.cfg.Logger()
		stack = service.GetCloudFormationStackName(env)
		cfn   = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
	)

	log.Infof("Deleting cloud formation stack '%s'", stack)

	exists, err := cfn.StackExists(stack)

	if err != nil {
		return nil
	}

	if !exists {
		log.Printf("  Stack '%s' does not exist\n", stack)
		return nil
	}

	return cfn.DeleteStack(stack)
}
