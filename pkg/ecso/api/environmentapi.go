package api

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type EnvironmentAPI interface {
	DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error)
	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool, w io.Writer) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment, w io.Writer) error
	IsEnvironmentUp(env *ecso.Environment) (bool, error)
	GetCurrentAWSAccount() (string, error)
	GetEcsoBucket(env *ecso.Environment) (string, error)
	GetECSServices(env *ecso.Environment) ([]*ecs.Service, error)
	GetECSTasks(env *ecso.Environment) ([]*ecs.Task, error)
	GetECSContainers(env *ecso.Environment) (ContainerList, error)
	SendNotification(env *ecso.Environment, msg string) error
}

// New creates a new API
func NewEnvironmentAPI(
	cloudformationAPI cloudformationiface.CloudFormationAPI,
	cloudwatchlogsAPI cloudwatchlogsiface.CloudWatchLogsAPI,
	ecsAPI ecsiface.ECSAPI,
	route53API route53iface.Route53API,
	s3API s3iface.S3API,
	snsAPI snsiface.SNSAPI,
	stsAPI stsiface.STSAPI,
) EnvironmentAPI {
	return &environmentAPI{
		cloudformationAPI: cloudformationAPI,
		cloudwatchlogsAPI: cloudwatchlogsAPI,
		ecsAPI:            ecsAPI,
		route53API:        route53API,
		s3API:             s3API,
		snsAPI:            snsAPI,
		stsAPI:            stsAPI,
	}
}

type environmentAPI struct {
	cloudformationAPI cloudformationiface.CloudFormationAPI
	cloudwatchlogsAPI cloudwatchlogsiface.CloudWatchLogsAPI
	ecsAPI            ecsiface.ECSAPI
	route53API        route53iface.Route53API
	s3API             s3iface.S3API
	snsAPI            snsiface.SNSAPI
	stsAPI            stsiface.STSAPI
}

func (api *environmentAPI) GetCurrentAWSAccount() (string, error) {
	resp, err := api.stsAPI.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *resp.Account, nil
}

func (api *environmentAPI) GetEcsoBucket(env *ecso.Environment) (string, error) {
	resp, err := api.stsAPI.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ecso-%s-%s", env.Region, *resp.Account), nil
}

func (api *environmentAPI) DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error) {
	var (
		stack       = env.GetCloudFormationStackName()
		cfnConsole  = util.CloudFormationConsoleURL(stack, env.Region)
		ecsConsole  = util.ClusterConsoleURL(env.GetClusterName(), env.Region)
		description = &EnvironmentDescription{Name: env.Name}
	)

	cfn := helpers.NewCloudFormationHelper(
		env.Region,
		api.cloudformationAPI,
		api.s3API,
		api.stsAPI)

	outputs, err := cfn.GetStackOutputs(stack)

	if err != nil {
		return description, err
	}

	description.CloudFormationOutputs = make(map[string]string)
	description.CloudFormationConsoleURL = cfnConsole
	description.ECSConsoleURL = ecsConsole
	description.CloudWatchLogsConsoleURL = util.CloudWatchLogsConsoleURL(outputs["LogGroup"], env.Region)
	description.ECSClusterBaseURL = fmt.Sprintf("http://%s", outputs["RecordSet"])

	for k, v := range outputs {
		description.CloudFormationOutputs[k] = v
	}

	return description, nil
}

func (api *environmentAPI) GetECSContainers(env *ecso.Environment) (ContainerList, error) {
	tasks, err := api.GetECSTasks(env)
	if err != nil {
		return nil, err
	}

	return LoadContainerList(tasks, api.ecsAPI)
}

func (api *environmentAPI) GetECSServices(env *ecso.Environment) ([]*ecs.Service, error) {
	var (
		count     = 0
		batchSize = 10
		batches   = make([][]*string, 0) // [][]service arns
		services  = make([]*ecs.Service, 0)
	)

	params := &ecs.ListServicesInput{
		Cluster: aws.String(env.GetClusterName()),
	}

	// TODO handle pages concurrently
	if err := api.ecsAPI.ListServicesPages(params, func(o *ecs.ListServicesOutput, last bool) bool {
		if count%batchSize == 0 {
			batches = append(batches, make([]*string, 0))
		}

		for i := range o.ServiceArns {
			batch := append(batches[len(batches)-1], o.ServiceArns[i])
			batches[len(batches)-1] = batch
			count = count + 1
		}

		return !last
	}); err != nil {
		return services, err
	}

	for _, batch := range batches {
		if len(batch) == 0 {
			continue
		}

		desc, err := api.ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
			Services: batch,
			Cluster:  aws.String(env.GetClusterName()),
		})

		if err != nil {
			return services, err
		}

		for _, svc := range desc.Services {
			services = append(services, svc)
		}
	}

	return services, nil
}

func (api *environmentAPI) GetECSTasks(env *ecso.Environment) ([]*ecs.Task, error) {
	taskArns := make([]*string, 0)

	params := &ecs.ListTasksInput{
		Cluster: aws.String(env.GetClusterName()),
	}

	if err := api.ecsAPI.ListTasksPages(params, func(o *ecs.ListTasksOutput, lastPage bool) bool {
		taskArns = append(taskArns, o.TaskArns...)
		return !lastPage
	}); err != nil {
		return nil, err
	}

	resp, err := api.ecsAPI.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(env.GetClusterName()),
		Tasks:   taskArns,
	})

	return resp.Tasks, err
}

func (api *environmentAPI) IsEnvironmentUp(env *ecso.Environment) (bool, error) {
	cfn := helpers.NewCloudFormationHelper(
		env.Region,
		api.cloudformationAPI,
		api.s3API,
		api.stsAPI)

	return cfn.StackExists(env.GetCloudFormationStackName())
}

func (api *environmentAPI) EnvironmentDown(p *ecso.Project, env *ecso.Environment, w io.Writer) error {
	var (
		cfnHelper      = helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)
		r53Helper      = helpers.NewRoute53Helper(api.route53API)
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
		datadogDNSName = fmt.Sprintf("%s.%s.%s", "datadog", env.GetClusterName(), zone)
		serviceAPI     = NewServiceAPI(api.cloudformationAPI, api.cloudwatchlogsAPI, api.ecsAPI, api.route53API, api.s3API, api.snsAPI, api.stsAPI)
		info           = ui.NewInfoWriter(w)
	)

	// TODO do these concurrently
	for _, service := range p.Services {
		if err := serviceAPI.ServiceDown(p, env, service, w); err != nil {
			return err
		}

		fmt.Fprint(w, "\n")
	}

	fmt.Fprintf(info, "Deleting environment Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnHelper.DeleteStack(env.GetCloudFormationStackName(), ui.NewPrefixWriter(w, "  ")); err != nil {
		return err
	}

	fmt.Fprint(w, "\n")
	fmt.Fprintf(info, "Deleting %s SRV records", datadogDNSName)

	return r53Helper.DeleteResourceRecordSetsByName(
		datadogDNSName,
		zone,
		"Deleted by ecso environment rm",
		ui.NewPrefixWriter(w, "  "))
}

func (api *environmentAPI) EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool, w io.Writer) error {
	info := ui.NewInfoWriter(w)
	version := util.VersionFromTime(time.Now())

	fmt.Fprintf(info, "Updating environment to version %s", version)

	bucket, err := api.GetEcsoBucket(env)
	if err != nil {
		return err
	}

	if err := api.uploadEnvironmentResources(bucket, env, version, w); err != nil {
		return err
	}

	result, err := api.deployEnvironmentStack(bucket, p, env, version, dryRun, w)
	if err != nil {
		return err
	}

	if dryRun {
		resp, err := api.cloudformationAPI.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(result.ChangeSetID),
			StackName:     aws.String(result.StackID),
		})

		if err != nil {
			return err
		}

		fmt.Fprintf(info, "\n%s", "The following changes were detected:")
		fmt.Fprintf(w, "\n%s\n", resp)
	}

	return nil
}

func (api *environmentAPI) SendNotification(env *ecso.Environment, msg string) error {
	var (
		stack = env.GetCloudFormationStackName()
	)

	cfn := helpers.NewCloudFormationHelper(env.Region, api.cloudformationAPI, api.s3API, api.stsAPI)

	outputs, err := cfn.GetStackOutputs(stack)
	if err != nil {
		return err
	}

	if topic, ok := outputs["NotificationsTopic"]; ok {
		_, err := api.snsAPI.Publish(&sns.PublishInput{
			Message: aws.String(msg),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"Environment": {
					DataType:    aws.String("String"),
					StringValue: aws.String(env.Name),
				},
			},
			TopicArn: aws.String(topic),
		})

		return err
	}

	return nil
}

func (api *environmentAPI) deployEnvironmentStack(bucket string, project *ecso.Project, env *ecso.Environment, version string, dryRun bool, w io.Writer) (*helpers.DeploymentResult, error) {
	var (
		stackName = env.GetCloudFormationStackName()
		prefix    = env.GetDeploymentBucketPrefix(version)
		template  = env.GetCloudFormationTemplateFile()
		tags      = env.CloudFormationTags
		params    = env.CloudFormationParameters
		info      = ui.NewInfoWriter(w)
	)

	fmt.Fprintf(info, "Deploying Cloud Formation stack for the '%s' environment", env.Name)

	cfn := helpers.NewCloudFormationHelper(
		env.Region,
		api.cloudformationAPI,
		api.s3API,
		api.stsAPI)

	if tags == nil {
		tags = make(map[string]string)
	}

	tags["ecso-cli-version"] = project.EcsoVersion
	tags["version"] = version

	params["S3BucketName"] = bucket
	params["Version"] = version
	params["S3KeyPrefix"] = env.GetBaseBucketPrefix()

	pkg, err := cfn.Package(template, bucket, prefix, tags, params, ui.NewPrefixWriter(w, "  "))
	if err != nil {
		return nil, err
	}

	return cfn.Deploy(pkg, stackName, dryRun, ui.NewPrefixWriter(w, "  "))
}

func (api *environmentAPI) uploadEnvironmentResources(bucket string, env *ecso.Environment, version string, w io.Writer) error {
	info := ui.NewInfoWriter(w)

	fmt.Fprintf(info, "Uploading resources for the '%s' environment to S3", env.Name)

	s3Helper := helpers.NewS3Helper(api.s3API, env.Region)

	return s3Helper.UploadDir(env.GetResourceDir(), bucket, env.GetResourceBucketPrefix(), ui.NewPrefixWriter(w, "  "))
}
