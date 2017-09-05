package api

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceAPI interface {
	DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error)
	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service, w io.Writer) (*ServiceDescription, error)
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service, w io.Writer) error
	ServiceEvents(p *ecso.Project, env *ecso.Environment, s *ecso.Service, f func(*ecs.ServiceEvent, error)) (cancel func(), err error)
	ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error)
	ServiceRollback(p *ecso.Project, env *ecso.Environment, s *ecso.Service, version string, w io.Writer) (*ServiceDescription, error)
	GetECSContainers(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ContainerList, error)
	GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error)
	GetECSTasks(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*ecs.Task, error)
	GetECSContainerImage(taskDefinitionArn, containerName string, env *ecso.Environment) (string, error)
	GetAvailableVersions(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ServiceVersionList, error)
}

// New creates a new API
func NewServiceAPI(
	cloudformationAPI cloudformationiface.CloudFormationAPI,
	cloudwatchlogsAPI cloudwatchlogsiface.CloudWatchLogsAPI,
	ecsAPI ecsiface.ECSAPI,
	route53API route53iface.Route53API,
	s3API s3iface.S3API,
	snsAPI snsiface.SNSAPI,
	stsAPI stsiface.STSAPI,
) ServiceAPI {
	return &serviceAPI{
		cloudformationAPI: cloudformationAPI,
		cloudwatchlogsAPI: cloudwatchlogsAPI,
		ecsAPI:            ecsAPI,
		route53API:        route53API,
		s3API:             s3API,
		snsAPI:            snsAPI,
		stsAPI:            stsAPI,
	}
}

type serviceAPI struct {
	cloudformationAPI cloudformationiface.CloudFormationAPI
	cloudwatchlogsAPI cloudwatchlogsiface.CloudWatchLogsAPI
	ecsAPI            ecsiface.ECSAPI
	route53API        route53iface.Route53API
	s3API             s3iface.S3API
	snsAPI            snsiface.SNSAPI
	stsAPI            stsiface.STSAPI
}

func (api *serviceAPI) GetECSContainers(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ContainerList, error) {
	tasks, err := api.GetECSTasks(p, env, s)
	if err != nil {
		return nil, err
	}

	return LoadContainerList(tasks, api.ecsAPI)
}

func (api *serviceAPI) GetAvailableVersions(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (ServiceVersionList, error) {
	envAPI := NewEnvironmentAPI(api.cloudformationAPI, api.cloudwatchlogsAPI, api.ecsAPI, api.route53API, api.s3API, api.snsAPI, api.stsAPI)

	bucket, err := envAPI.GetEcsoBucket(env)
	if err != nil {
		return nil, err
	}

	prefix := s.GetDeploymentBucketPrefix(env)

	resp, err := api.s3API.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, err
	}

	versions := make([]*ServiceVersion, 0)
	labels := make(map[string]bool)

	for _, o := range resp.Contents {
		suffix := strings.TrimPrefix(*o.Key, prefix+"/")
		tokens := strings.Split(suffix, "/")
		labels[tokens[0]] = true
	}

	for k := range labels {
		versions = append(versions, &ServiceVersion{
			Service: s.Name,
			Label:   k,
		})
	}

	return ServiceVersionList(versions), nil
}

func (api *serviceAPI) GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error) {
	var (
		cfn = helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
	)

	outputs, err := cfn.GetStackOutputs(s.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	if serviceName, ok := outputs["Service"]; ok {
		resp, err := api.ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
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
	resp, err := api.ecsAPI.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
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

	runningService, err := api.GetECSService(p, env, s)

	if err != nil || runningService == nil {
		return result, err
	}

	tasks, err := api.ecsAPI.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(env.GetClusterName()),
		ServiceName: runningService.ServiceName,
	})

	if err != nil {
		return result, err
	}

	resp, err := api.ecsAPI.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(env.GetClusterName()),
		Tasks:   tasks.TaskArns,
	})

	if err != nil {
		return result, err
	}

	return resp.Tasks, nil
}

func (api *serviceAPI) DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error) {
	cfn := helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)

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
		CloudWatchLogsConsoleURL: util.CloudWatchLogsConsoleURL(service.GetCloudWatchLogGroup(env), env.Region),
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

func (api *serviceAPI) ServiceDown(project *ecso.Project, env *ecso.Environment, service *ecso.Service, w io.Writer) error {
	if err := api.deleteServiceStack(env, service, w); err != nil {
		return err
	}

	fmt.Fprint(w, "\n")

	if err := api.clearServiceDNSRecords(env, service, w); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) ServiceEvents(p *ecso.Project, env *ecso.Environment, s *ecso.Service, f func(*ecs.ServiceEvent, error)) (cancel func(), err error) {
	runningService, err := api.GetECSService(p, env, s)
	if err != nil {
		return nil, err
	}

	if runningService == nil {
		return nil, fmt.Errorf("No service named %s is running", s.Name)
	}

	h := helpers.NewECSHelper(api.ecsAPI)

	return h.LogServiceEvents(*runningService.ServiceArn, env.GetClusterName(), f), nil
}

func (api *serviceAPI) ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	streams, err := api.cloudwatchlogsAPI.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(s.GetCloudWatchLogGroup(env)),
		LogStreamNamePrefix: aws.String(s.GetCloudWatchLogStreamPrefix(env)),
	})

	if err != nil {
		return nil, err
	}

	streamNames := make([]*string, 0)

	for _, stream := range streams.LogStreams {
		streamNames = append(streamNames, stream.LogStreamName)
	}

	resp, err := api.cloudwatchlogsAPI.FilterLogEvents(&cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:   aws.String(s.GetCloudWatchLogGroup(env)),
		Interleaved:    aws.Bool(true),
		LogStreamNames: streamNames,
	})

	if err != nil {
		return nil, err
	}

	return resp.Events, nil
}

func (api *serviceAPI) ServiceRollback(project *ecso.Project, env *ecso.Environment, service *ecso.Service, version string, w io.Writer) (*ServiceDescription, error) {
	envAPI := NewEnvironmentAPI(api.cloudformationAPI, api.cloudwatchlogsAPI, api.ecsAPI, api.route53API, api.s3API, api.snsAPI, api.stsAPI)

	bucket, err := envAPI.GetEcsoBucket(env)
	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
	pkg := helpers.NewPackage(bucket, service.GetDeploymentBucketPrefixForVersion(env, version), env.Region)

	exists, err := cfn.PackageIsUploadedToS3(pkg)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("Version %s of service %s not found", version, service.Name)
	}

	if err := envAPI.SendNotification(env, fmt.Sprintf("Commenced rollback of %s version %s to %s", service.Name, version, env.Name)); err != nil {
		fmt.Fprintf(w, "WARNING Failed to send rollback commencing notification to sns. %s", err.Error())
	}

	// deploy the service cfn stack
	if err := api.deployServiceStack(pkg, env, service, w); err != nil {
		if err := envAPI.SendNotification(env, fmt.Sprintf("Failed to deploy %s to %s", service.Name, env.Name)); err != nil {
			fmt.Fprintf(w, "WARNING Failed to send deployment failure notification to sns. %s", err.Error())
		}
		return nil, err
	}

	if err := envAPI.SendNotification(env, fmt.Sprintf("Completed rollback of %s version %s to %s", service.Name, version, env.Name)); err != nil {
		fmt.Fprintf(w, "WARNING Failed to send deployment completed notification to sns. %s", err.Error())
	}

	return api.DescribeService(env, service)
}

func (api *serviceAPI) ServiceUp(project *ecso.Project, env *ecso.Environment, service *ecso.Service, w io.Writer) (*ServiceDescription, error) {
	version := util.VersionFromTime(time.Now())
	envAPI := NewEnvironmentAPI(api.cloudformationAPI, api.cloudwatchlogsAPI, api.ecsAPI, api.route53API, api.s3API, api.snsAPI, api.stsAPI)

	bucket, err := envAPI.GetEcsoBucket(env)
	if err != nil {
		return nil, err
	}

	if err := envAPI.SendNotification(env, fmt.Sprintf("Commenced deployment of %s to %s", service.Name, env.Name)); err != nil {
		fmt.Fprintf(w, "WARNING Failed to send deployment commencing notification to sns. %s", err.Error())
	}

	// register task
	taskDefinition, err := api.registerECSTaskDefinition(project, env, service, w)
	if err != nil {
		if err := envAPI.SendNotification(env, fmt.Sprintf("Failed to deploy %s to %s", service.Name, env.Name)); err != nil {
			fmt.Fprintf(w, "WARNING Failed to send deployment failure notification to sns. %s", err.Error())
		}
		return nil, err
	}

	// deploy the service cfn stack
	if err := api.packageAndDeployServiceStack(bucket, project, env, service, taskDefinition, version, w); err != nil {
		if err := envAPI.SendNotification(env, fmt.Sprintf("Failed to deploy %s to %s", service.Name, env.Name)); err != nil {
			fmt.Fprintf(w, "WARNING Failed to send deployment failure notification to sns. %s", err.Error())
		}
		return nil, err
	}

	if err := envAPI.SendNotification(env, fmt.Sprintf("Completed deployment of %s to %s", service.Name, env.Name)); err != nil {
		fmt.Fprintf(w, "WARNING Failed to send deployment completed notification to sns. %s", err.Error())
	}

	return api.DescribeService(env, service)
}

func (api *serviceAPI) deployServiceStack(pkg *helpers.Package, env *ecso.Environment, service *ecso.Service, w io.Writer) error {
	var (
		stackName = service.GetCloudFormationStackName(env)
		info      = ui.NewInfoWriter(w)
		cfn       = helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
	)

	fmt.Fprintf(info, "Deploying service cloudformation stack '%s'...", stackName)

	result, err := cfn.Deploy(pkg, stackName, false, ui.NewPrefixWriter(w, "  "))
	if err != nil {
		return err
	}

	if !result.DidRequireUpdating {
		fmt.Fprintf(w, "  No updates were required to Cloud Formation stack '%s'\n", result.StackID)
	}

	return nil
}

func (api *serviceAPI) packageAndDeployServiceStack(bucket string, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition, version string, w io.Writer) error {
	var (
		prefix   = service.GetDeploymentBucketPrefixForVersion(env, version)
		template = service.GetCloudFormationTemplateFile()
		cfn      = helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
	)

	params, err := getServiceStackParameters(cfn, project, env, service, taskDefinition, version)

	if err != nil {
		return err
	}

	tags := getServiceStackTags(project, env, service, version)

	pkg, err := cfn.Package(template, bucket, prefix, tags, params, ui.NewPrefixWriter(w, "  "))
	if err != nil {
		return err
	}

	return api.deployServiceStack(pkg, env, service, w)
}

func getServiceStackParameters(cfn helpers.CloudFormationHelper, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition, version string) (map[string]string, error) {

	outputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"Cluster":        outputs["Cluster"],
		"AlertsTopic":    outputs["AlertsTopic"],
		"Version":        version,
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

func getServiceStackTags(project *ecso.Project, env *ecso.Environment, service *ecso.Service, version string) map[string]string {
	tags := map[string]string{
		"project":          project.Name,
		"environment":      env.Name,
		"ecso-cli-version": project.EcsoVersion,
		"version":          version,
	}

	for k, v := range service.Tags {
		tags[k] = v
	}

	return tags
}

func (api *serviceAPI) registerECSTaskDefinition(project *ecso.Project, env *ecso.Environment, service *ecso.Service, w io.Writer) (*ecs.TaskDefinition, error) {
	var (
		taskName = service.GetECSTaskDefinitionName(env)
		info     = ui.NewInfoWriter(w)
	)

	// TODO: fully qualify the path to the service compose file
	// taskDefinition, err := ConvertToTaskDefinition(taskName, service.ComposeFile)
	fmt.Fprintf(info, "Converting '%s' to task definition...\n", service.ComposeFile)

	taskDefinition, err := service.GetECSTaskDefinition(env)
	if err != nil {
		return nil, err
	}

	for _, container := range taskDefinition.ContainerDefinitions {
		fmt.Fprintf(info, "Configuring cloudwatch logs for %s container\n", *container.Name)
		container.SetLogConfiguration(&ecs.LogConfiguration{
			LogDriver: aws.String(ecs.LogDriverAwslogs),
			Options: map[string]*string{
				"awslogs-region":        aws.String(env.Region),
				"awslogs-group":         aws.String(service.GetCloudWatchLogGroup(env)),
				"awslogs-stream-prefix": aws.String(service.GetCloudWatchLogStreamPrefix(env)),
			},
		})

		// TODO probably don't automatically add service discovery env config
		//      if people want to use service discover they can just add the
		//      required env vars to the helpers docker-compose file. This is
		//      less magic and more flexible
		for _, p := range container.PortMappings {
			fmt.Fprintf(info, "Adding service discovery env var SERVICE_%d_NAME to %s container\n", *p.ContainerPort, *container.Name)
			container.Environment = append(container.Environment, &ecs.KeyValuePair{
				Name:  aws.String(fmt.Sprintf("SERVICE_%d_NAME", *p.ContainerPort)),
				Value: aws.String(fmt.Sprintf("%s.%s", service.Name, env.GetClusterName())),
			})
		}
	}

	fmt.Fprintf(info, "Registering ECS task definition '%s'...", taskName)
	resp, err := api.ecsAPI.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
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

	fmt.Fprintf(
		w,
		"  Registered ECS task definition %s:%d\n\n",
		*resp.TaskDefinition.Family,
		*resp.TaskDefinition.Revision)

	return resp.TaskDefinition, nil
}

func (api *serviceAPI) clearServiceDNSRecords(env *ecso.Environment, service *ecso.Service, w io.Writer) error {
	var (
		r53Helper = helpers.NewRoute53Helper(api.route53API)
		dnsName   = fmt.Sprintf("%s.%s.", service.Name, env.CloudFormationParameters["DNSZone"])
		info      = ui.NewInfoWriter(w)
	)

	fmt.Fprintf(info, "Deleting any service SRV DNS records for %s...", dnsName)

	if err := r53Helper.DeleteResourceRecordSetsByName(dnsName, env.CloudFormationParameters["DNSZone"], "Deleted by ecso service down", ui.NewPrefixWriter(w, "  ")); err != nil {
		return err
	}

	return nil
}

func (api *serviceAPI) deleteServiceStack(env *ecso.Environment, service *ecso.Service, w io.Writer) error {
	var (
		stack = service.GetCloudFormationStackName(env)
		cfn   = helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
		info  = ui.NewInfoWriter(w)
	)

	fmt.Fprintf(info, "Deleting cloud formation stack '%s'", stack)

	exists, err := cfn.StackExists(stack)

	if err != nil {
		return nil
	}

	if !exists {
		fmt.Fprintf(w, "  Stack '%s' does not exist\n", stack)
		return nil
	}

	return cfn.DeleteStack(stack, ui.NewPrefixWriter(w, "  "))
}
