package api

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type EnvironmentAPI interface {
	DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error)
	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment) error
	IsEnvironmentUp(env *ecso.Environment) (bool, error)
	SendNotification(env *ecso.Environment, msg string) error
	GetECSServices(env *ecso.Environment) ([]*ecs.Service, error)
	GetECSTasks(env *ecso.Environment) ([]*ecs.Task, error)
	GetECSContainers(env *ecso.Environment) (ContainerList, error)
}

// New creates a new API
func NewEnvironmentAPI(log log.Logger, registryFactory awsregistry.RegistryFactory) EnvironmentAPI {
	return &environmentAPI{
		log:             log,
		registryFactory: registryFactory,
	}
}

type environmentAPI struct {
	log             log.Logger
	registryFactory awsregistry.RegistryFactory
}

type EnvironmentDescription struct {
	Name                     string
	CloudFormationConsoleURL string
	CloudWatchLogsConsoleURL string
	ECSConsoleURL            string
	ECSClusterBaseURL        string
	CloudFormationOutputs    map[string]string
}

func (api *environmentAPI) DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error) {
	var (
		stack       = env.GetCloudFormationStackName()
		cfnConsole  = util.CloudFormationConsoleURL(stack, env.Region)
		ecsConsole  = util.ClusterConsoleURL(env.GetClusterName(), env.Region)
		description = &EnvironmentDescription{Name: env.Name}
	)

	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return description, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())

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
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	tasks, err := api.GetECSTasks(env)

	if err != nil {
		return nil, err
	}

	return LoadContainerList(tasks, reg.ECSAPI())
}

func (api *environmentAPI) GetECSServices(env *ecso.Environment) ([]*ecs.Service, error) {
	var (
		count     = 0
		batchSize = 10
		batches   = make([][]*string, 0) // [][]service arns
		services  = make([]*ecs.Service, 0)
	)

	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	params := &ecs.ListServicesInput{
		Cluster: aws.String(env.GetClusterName()),
	}

	ecsAPI := reg.ECSAPI()

	// TODO handle pages concurrently
	if err := ecsAPI.ListServicesPages(params, func(o *ecs.ListServicesOutput, last bool) bool {
		if count%batchSize == 0 {
			batches = append(batches, make([]*string, 0))
		}

		for i, _ := range o.ServiceArns {
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

		desc, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
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
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return nil, err
	}

	ecsAPI := reg.ECSAPI()
	params := &ecs.ListTasksInput{
		Cluster: aws.String(env.GetClusterName()),
	}

	err = ecsAPI.ListTasksPages(params, func(o *ecs.ListTasksOutput, lastPage bool) bool {
		taskArns = append(taskArns, o.TaskArns...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	resp, err := ecsAPI.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(env.GetClusterName()),
		Tasks:   taskArns,
	})

	return resp.Tasks, err
}

func (api *environmentAPI) IsEnvironmentUp(env *ecso.Environment) (bool, error) {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return false, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())

	return cfn.StackExists(env.GetCloudFormationStackName())
}

func (api *environmentAPI) EnvironmentDown(p *ecso.Project, env *ecso.Environment) error {
	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return err
	}

	var (
		cfnHelper      = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())
		r53Helper      = helpers.NewRoute53Helper(reg.Route53API(), api.log.Child())
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
		datadogDNSName = fmt.Sprintf("%s.%s.%s", "datadog", env.GetClusterName(), zone)
		serviceAPI     = NewServiceAPI(api.log, api.registryFactory)
	)

	// TODO do these concurrently
	for _, service := range p.Services {
		if err := serviceAPI.ServiceDown(p, env, service); err != nil {
			return err
		}
		api.log.Printf("\n")
	}

	api.log.Infof("Deleting environment Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnHelper.DeleteStack(env.GetCloudFormationStackName()); err != nil {
		return err
	}

	api.log.Printf("\n")
	api.log.Infof("Deleting %s SRV records", datadogDNSName)

	return r53Helper.DeleteResourceRecordSetsByName(
		datadogDNSName,
		zone,
		"Deleted by ecso environment rm")
}

func (api *environmentAPI) EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error {
	var (
		version  = util.VersionFromTime(time.Now())
		stack    = env.GetCloudFormationStackName()
		template = env.GetCloudFormationTemplateFile()
		prefix   = env.GetBaseBucketPrefix(version)
		tags     = env.CloudFormationTags
		params   = env.CloudFormationParameters
	)

	reg, err := api.registryFactory.ForRegion(env.Region)
	if err != nil {
		return err
	}

	stsClient := reg.STSAPI()
	s3Helper := helpers.NewS3Helper(reg.S3API(), env.Region, api.log.Child())

	bucket, err := util.GetEcsoBucket(stsClient, env.Region)
	if err != nil {
		return err
	}

	api.log.Infof("Uploading resources for the '%s' environment to S3", env.Name)

	if err := s3Helper.UploadDir(env.GetResourceDir(), bucket, env.GetResourceBucketPrefix(version)); err != nil {
		return err
	}

	api.log.Printf("\n")
	api.log.Infof("Deploying Cloud Formation stack for the '%s' environment", env.Name)

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())

	if tags == nil {
		tags = make(map[string]string)
	}

	tags["ecso-version"] = p.EcsoVersion
	params["S3BucketName"] = bucket
	params["S3KeyPrefix"] = env.GetBaseBucketPrefix(version)

	result, err := cfn.PackageAndDeploy(stack, template, prefix, tags, params, dryRun)

	if dryRun {
		cfnAPI := reg.CloudFormationAPI()

		resp, err := cfnAPI.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(result.ChangeSetID),
			StackName:     aws.String(result.StackID),
		})

		if err != nil {
			return err
		}

		api.log.Printf("\n")
		api.log.Infof("The following changes were detected:")
		api.log.Printf("\n%s\n", resp)
	}

	return err
}

func (api *environmentAPI) SendNotification(env *ecso.Environment, msg string) error {
	var (
		stack = env.GetCloudFormationStackName()
	)

	reg, err := api.registryFactory.ForRegion(env.Region)

	if err != nil {
		return err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.log.Child())

	outputs, err := cfn.GetStackOutputs(stack)

	if topic, ok := outputs["NotificationsTopic"]; ok {
		snsAPI := reg.SNSAPI()

		_, err := snsAPI.Publish(&sns.PublishInput{
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
