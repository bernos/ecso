package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
)

func (api *api) EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error {
	var (
		log      = api.cfg.Logger()
		stack    = env.GetCloudFormationStackName()
		template = env.GetCloudFormationTemplateFile()
		prefix   = env.GetCloudFormationBucketPrefix()
		tags     = env.CloudFormationTags
		params   = env.CloudFormationParameters
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	log.Infof("Deploying Cloud Formation stack for the '%s' environment", env.Name)

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child().Printf)
	exists, err := cfn.StackExists(stack)

	if err != nil {
		return err
	}

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
