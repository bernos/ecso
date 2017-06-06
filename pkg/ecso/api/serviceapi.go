package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceAPI interface {
	DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error)
	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ServiceDescription, error)
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceEvents(p *ecso.Project, env *ecso.Environment, s *ecso.Service, f func(*ecs.ServiceEvent, error)) (cancel func(), err error)
	ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error)
	GetECSContainers(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ContainerList, error)
	GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error)
	GetECSTasks(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*ecs.Task, error)

	GetECSContainerImage(taskDefinitionArn, containerName string, env *ecso.Environment) (string, error)
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
func NewServiceAPI(log log.Logger, registryFactory awsregistry.RegistryFactory) ServiceAPI {
	return &serviceAPI{
		log:             log,
		registryFactory: registryFactory,
	}
}

type serviceAPI struct {
	log             log.Logger
	registryFactory awsregistry.RegistryFactory
}

func (api *serviceAPI) GetECSContainers(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ContainerList, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	tasks, err := api.GetECSTasks(p, env, s)

	if err != nil {
		return nil, err
	}

	return LoadContainerList(tasks, reg.ECSAPI())
}

func (api *serviceAPI) GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	var (
		cfn    = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())
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

func (api *serviceAPI) GetECSContainerImage(taskDefinitionArn, containerName string, env *ecso.Environment) (string, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return "", err
	}

	ecsAPI := reg.ECSAPI()

	resp, err := ecsAPI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionArn),
	})

	if err != nil {
		return "", err
	}

	for _, c := range resp.TaskDefinition.ContainerDefinitions {
		if *c.Name == containerName {
			return *c.Image, nil
		}
	}

	return "", nil
}

func (api *serviceAPI) GetECSTasks(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*ecs.Task, error) {
	result := make([]*ecs.Task, 0)
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return result, err
	}

	runningService, err := api.GetECSService(p, env, s)

	if err != nil || runningService == nil {
		return result, err
	}

	ecsAPI := reg.ECSAPI()

	tasks, err := ecsAPI.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(env.GetClusterName()),
		ServiceName: runningService.ServiceName,
	})

	if err != nil {
		return result, err
	}

	resp, err := ecsAPI.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(env.GetClusterName()),
		Tasks:   tasks.TaskArns,
	})

	if err != nil {
		return result, err
	}

	return resp.Tasks, nil
}

func (api *serviceAPI) DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())

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
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return err
	}

	if err := api.deleteServiceStack(reg, env, service); err != nil {
		return err
	}

	api.log.Printf("\n")

	if err := api.clearServiceDNSRecords(reg, env, service); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) ServiceEvents(p *ecso.Project, env *ecso.Environment, s *ecso.Service, f func(*ecs.ServiceEvent, error)) (cancel func(), err error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	runningService, err := api.GetECSService(p, env, s)

	if err != nil {
		return nil, err
	}

	if runningService == nil {
		return nil, fmt.Errorf("No service named %s is running", s.Name)
	}

	h := helpers.NewECSHelper(reg.ECSAPI(), api.log.Child())

	return h.LogServiceEvents(*runningService.ServiceArn, env.GetClusterName(), f), nil
}

func (api *serviceAPI) ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	cwLogsAPI := reg.CloudWatchLogsAPI()

	streams, err := cwLogsAPI.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(s.GetCloudWatchLogGroupName(env)),
		LogStreamNamePrefix: aws.String(s.GetCloudWatchLogStreamPrefix(env)),
	})

	if err != nil {
		return nil, err
	}

	streamNames := make([]*string, 0)

	for _, stream := range streams.LogStreams {
		streamNames = append(streamNames, stream.LogStreamName)
	}

	resp, err := cwLogsAPI.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   aws.String(s.GetCloudWatchLogGroupName(env)),
		Interleaved:    aws.Bool(true),
		LogStreamNames: streamNames,
	})

	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}

func (api *serviceAPI) ServiceUp(project *ecso.Project, env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	// set env vars so that they are available when converting the docker
	// compose file to a task definition
	if err := api.setEnv(project, env, service); err != nil {
		return nil, err
	}

	envAPI := NewEnvironmentAPI(api.log, api.registryFactory)

	if err := envAPI.SendNotification(env, fmt.Sprintf("Commenced deployment of %s to %s", service.Name, env.Name)); err != nil {
		api.log.Printf("WARNING Failed to send deployment commencing notification to sns. %s", err.Error())
	}

	// register task
	taskDefinition, err := api.registerECSTaskDefinition(reg, project, env, service)

	if err != nil {
		if err := envAPI.SendNotification(env, fmt.Sprintf("Failed to deploy %s to %s", service.Name, env.Name)); err != nil {
			api.log.Printf("WARNING Failed to send deployment failure notification to sns. %s", err.Error())
		}
		return nil, err
	}

	// deploy the service cfn stack
	if err := api.deployServiceStack(reg, project, env, service, taskDefinition); err != nil {
		if err := envAPI.SendNotification(env, fmt.Sprintf("Failed to deploy %s to %s", service.Name, env.Name)); err != nil {
			api.log.Printf("WARNING Failed to send deployment failure notification to sns. %s", err.Error())
		}
		return nil, err
	}

	if err := envAPI.SendNotification(env, fmt.Sprintf("Completed deployment of %s to %s", service.Name, env.Name)); err != nil {
		api.log.Printf("WARNING Failed to send deployment completed notification to sns. %s", err.Error())
	}

	return api.DescribeService(env, service)
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

func (api *serviceAPI) deployServiceStack(reg awsregistry.Registry, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition) error {
	var (
		stackName = service.GetCloudFormationStackName(env)
		prefix    = service.GetCloudFormationBucketPrefix(env)
		template  = service.GetCloudFormationTemplateFile()
		cfn       = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())
	)

	params, err := getServiceStackParameters(cfn, project, env, service, taskDefinition)

	if err != nil {
		return err
	}

	tags := getServiceStackTags(project, env, service)

	api.log.Infof("Deploying service cloudformation stack '%s'...", stackName)

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
		api.log.Printf("  No updates were required to Cloud Formation stack '%s'\n", result.StackID)
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

func (api *serviceAPI) registerECSTaskDefinition(reg awsregistry.Registry, project *ecso.Project, env *ecso.Environment, service *ecso.Service) (*ecs.TaskDefinition, error) {
	var (
		taskName  = service.GetECSTaskDefinitionName(env)
		ecsClient = reg.ECSAPI()
	)

	// TODO: fully qualify the path to the service compose file
	// taskDefinition, err := ConvertToTaskDefinition(taskName, service.ComposeFile)
	api.log.Infof("Converting '%s' to task definition...", service.ComposeFile)

	taskDefinition, err := service.GetECSTaskDefinition(env)

	api.log.Printf("\n")
	api.log.Infof("Registering ECS task definition '%s'...", taskName)

	if err != nil {
		return nil, err
	}

	for _, container := range taskDefinition.ContainerDefinitions {
		container.SetLogConfiguration(&ecs.LogConfiguration{
			LogDriver: aws.String(ecs.LogDriverAwslogs),
			Options: map[string]*string{
				"awslogs-region":        aws.String(env.Region),
				"awslogs-group":         aws.String(service.GetCloudWatchLogGroupName(env)),
				"awslogs-stream-prefix": aws.String(service.GetCloudWatchLogStreamPrefix(env)),
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

	api.log.Printf(
		"  Registered ECS task definition %s:%d\n\n",
		*resp.TaskDefinition.Family,
		*resp.TaskDefinition.Revision)

	return resp.TaskDefinition, nil
}
func (api *serviceAPI) clearServiceDNSRecords(reg awsregistry.Registry, env *ecso.Environment, service *ecso.Service) error {
	var (
		r53Helper = helpers.NewRoute53Helper(reg.Route53API(), api.log.Child())
		dnsName   = fmt.Sprintf("%s.%s.", service.Name, env.CloudFormationParameters["DNSZone"])
	)

	api.log.Infof("Deleting any service SRV DNS records for %s...", dnsName)

	if err := r53Helper.DeleteResourceRecordSetsByName(dnsName, env.CloudFormationParameters["DNSZone"], "Deleted by ecso service down"); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) deleteServiceStack(reg awsregistry.Registry, env *ecso.Environment, service *ecso.Service) error {
	var (
		stack = service.GetCloudFormationStackName(env)
		cfn   = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())
	)

	api.log.Infof("Deleting cloud formation stack '%s'", stack)

	exists, err := cfn.StackExists(stack)

	if err != nil {
		return nil
	}

	if !exists {
		api.log.Printf("  Stack '%s' does not exist\n", stack)
		return nil
	}

	return cfn.DeleteStack(stack)
}
