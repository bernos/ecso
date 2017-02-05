package api

import "github.com/bernos/ecso/pkg/ecso"

func (api *api) EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error {
	var (
		log      = api.cfg.Logger
		stack    = env.GetCloudFormationStackName()
		template = env.GetCloudFormationTemplateFile()
		prefix   = env.GetCloudFormationBucketPrefix()
		bucket   = env.CloudFormationBucket
		tags     = env.CloudFormationTags
		params   = env.CloudFormationParameters
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	log.Infof("Deploying Cloud Formation stack for the '%s' environment", env.Name)

	cfn := reg.CloudFormationService(log.PrefixPrintf("  "))
	exists, err := cfn.StackExists(stack)

	if err != nil {
		return err
	}

	if exists {
		_, err = cfn.PackageAndDeploy(stack, template, bucket, prefix, tags, params, dryRun)
	} else {
		_, err = cfn.PackageAndCreate(stack, template, bucket, prefix, tags, params, dryRun)
	}

	return err
}
