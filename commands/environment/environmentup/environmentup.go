package environmentup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/services"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type Options struct {
	EnvironmentName string
	DryRun          bool
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &envUpCommand{
		options: o,
	}
}

type logfn func(string, ...interface{})

type envUpCommand struct {
	options *Options
}

func (cmd *envUpCommand) Execute(ctx *ecso.CommandContext) error {

	if err := validateOptions(ctx, cmd.options); err != nil {
		return err
	}

	var (
		project  = ctx.Project
		cfg      = ctx.Config
		log      = cfg.Logger
		env      = project.Environments[cmd.options.EnvironmentName]
		registry = cfg.MustGetAWSClientRegistry(env.Region)
		cfn      = registry.CloudFormationService(log.PrefixPrintf("  "))
	)

	cfg.Logger.BannerBlue("Bringing up environment '%s'", env.Name)

	if cmd.options.DryRun {
		cfg.Logger.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := ensureTemplates(env, cfg.Logger.Infof); err != nil {
		return err
	}

	cfg.Logger.Infof("Deploying infrastructure Cloud Formation stack")

	result, err := deployStack(ctx, env, cmd.options.DryRun)

	if err != nil {
		return err
	}

	if cmd.options.DryRun {
		cfg.Logger.BannerGreen("Review the above changes and re-run the command without the --dry-run option to apply them")

		return nil
	}

	cfg.Logger.BannerGreen("Environment '%s' is up and running", env.Name)

	cfg.Logger.Dt("Cloud Formation stack", util.CloudFormationConsoleURL(result.StackID, env.Region))
	cfg.Logger.Dt("ECS Console", util.ClusterConsoleURL(env.GetClusterName(), env.Region))

	return cfn.LogStackOutputs(env.GetCloudFormationStackName(), cfg.Logger.Dt)
}

func deployStack(ctx *ecso.CommandContext, env *ecso.Environment, dryRun bool) (*services.DeploymentResult, error) {
	var (
		cfg        = ctx.Config
		stackName  = env.GetCloudFormationStackName()
		template   = env.GetCloudFormationTemplateFile()
		prefix     = env.GetCloudFormationBucketPrefix()
		bucket     = env.CloudFormationBucket
		params     = env.CloudFormationParameters
		tags       = env.CloudFormationTags
		registry   = cfg.MustGetAWSClientRegistry(env.Region)
		cfnService = registry.CloudFormationService(cfg.Logger.PrefixPrintf("  "))
	)

	return cfnService.PackageAndDeploy(stackName, template, bucket, prefix, tags, params, dryRun)
}

func ensureTemplates(env *ecso.Environment, log logfn) error {
	dst := env.GetCloudFormationTemplateDir()

	exists, err := util.DirExists(dst)

	if err != nil || exists {
		return err
	}

	return createCloudFormationTemplates(dst, log)
}

func createCloudFormationTemplates(dst string, log logfn) error {
	log("Copying infrastructure stack templates to %s", dst)

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	for file, content := range templates {
		if err := ioutil.WriteFile(filepath.Join(dst, file), []byte(content), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func validateOptions(ctx *ecso.CommandContext, opt *Options) error {
	if opt.EnvironmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(opt.EnvironmentName) {
		return fmt.Errorf("No environment named '%s' was found", opt.EnvironmentName)
	}

	return nil
}
