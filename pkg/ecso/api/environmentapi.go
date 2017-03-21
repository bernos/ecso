package api

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type EnvironmentAPI interface {
	DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error)
	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment) error
	IsEnvironmentUp(env *ecso.Environment) (bool, error)
	SendNotification(env *ecso.Environment, msg string) error
}

// New creates a new API
func NewEnvironmentAPI(cfg *ecso.Config) EnvironmentAPI {
	return &environmentAPI{cfg}
}

type environmentAPI struct {
	cfg *ecso.Config
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
		log         = api.cfg.Logger()
		stack       = env.GetCloudFormationStackName()
		cfnConsole  = util.CloudFormationConsoleURL(stack, env.Region)
		ecsConsole  = util.ClusterConsoleURL(env.GetClusterName(), env.Region)
		description = &EnvironmentDescription{Name: env.Name}
	)

	reg, err := awsregistry.ForRegion(env.Region)

	if err != nil {
		return description, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())

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

func (api *environmentAPI) IsEnvironmentUp(env *ecso.Environment) (bool, error) {
	reg, err := awsregistry.ForRegion(env.Region)

	if err != nil {
		return false, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), api.cfg.Logger().Child())

	return cfn.StackExists(env.GetCloudFormationStackName())
}

func (api *environmentAPI) EnvironmentDown(p *ecso.Project, env *ecso.Environment) error {
	reg, err := awsregistry.ForRegion(env.Region)

	if err != nil {
		return err
	}

	var (
		log            = api.cfg.Logger()
		cfnHelper      = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
		r53Helper      = helpers.NewRoute53Helper(reg.Route53API(), log.Child())
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
		datadogDNSName = fmt.Sprintf("%s.%s.%s", "datadog", env.GetClusterName(), zone)
		serviceAPI     = NewServiceAPI(api.cfg)
	)

	for _, service := range p.Services {
		if err := serviceAPI.ServiceDown(p, env, service); err != nil {
			return err
		}
		log.Printf("\n")
	}

	log.Infof("Deleting environment Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnHelper.DeleteStack(env.GetCloudFormationStackName()); err != nil {
		return err
	}

	log.Printf("\n")
	log.Infof("Deleting %s SRV records", datadogDNSName)

	return r53Helper.DeleteResourceRecordSetsByName(
		datadogDNSName,
		zone,
		"Deleted by ecso environment rm")
}

func (api *environmentAPI) EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error {
	var (
		log      = api.cfg.Logger()
		stack    = env.GetCloudFormationStackName()
		template = env.GetCloudFormationTemplateFile()
		prefix   = env.GetCloudFormationBucketPrefix()
		tags     = env.CloudFormationTags
		params   = env.CloudFormationParameters
	)

	reg, err := awsregistry.ForRegion(env.Region)

	if err != nil {
		return err
	}

	log.Infof("Deploying Cloud Formation stack for the '%s' environment", env.Name)

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
	exists, err := cfn.StackExists(stack)

	if err != nil {
		return err
	}

	if tags == nil {
		tags = make(map[string]string)
	}

	tags["ecso-version"] = p.EcsoVersion

	var result *helpers.DeploymentResult

	if exists {
		result, err = cfn.PackageAndDeploy(stack, template, prefix, tags, params, dryRun)
	} else {
		result, err = cfn.PackageAndCreate(stack, template, prefix, tags, params, dryRun)
	}

	if dryRun {
		cfnAPI := reg.CloudFormationAPI()

		resp, err := cfnAPI.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(result.ChangeSetID),
			StackName:     aws.String(result.StackID),
		})

		if err != nil {
			return err
		}

		log.Printf("\n")
		log.Infof("The following changes were detected:")
		log.Printf("\n%s\n", resp)
	}

	return err
}

func (api *environmentAPI) SendNotification(env *ecso.Environment, msg string) error {
	var (
		log   = api.cfg.Logger()
		stack = env.GetCloudFormationStackName()
	)

	reg, err := awsregistry.ForRegion(env.Region)

	if err != nil {
		return err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())

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
