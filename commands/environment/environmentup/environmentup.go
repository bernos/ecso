package environmentup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
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
		project = ctx.Project
		cfg     = ctx.Config
		env     = project.Environments[cmd.options.EnvironmentName]
	)

	cfn, err := ctx.Config.CloudFormationService(env.Region)

	if err != nil {
		return err
	}

	cfg.Logger.BannerBlue("Bringing up environment '%s'", env.Name)

	if cmd.options.DryRun {
		cfg.Logger.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := ensureTemplates(env, cfg.Logger.Infof); err != nil {
		return err
	}

	cfg.Logger.Infof("Packaging infrastructure stack templates")

	if err := deployStack(ctx, env, cmd.options.DryRun); err != nil {
		return err
	}

	if cmd.options.DryRun {
		cfg.Logger.BannerGreen("Review the above changes and re-run the command without the --dry-run option to apply them")

		return nil
	} else {
		cfg.Logger.BannerGreen("Environment '%s' is up and running", env.Name)

		return cfn.LogStackOutputs(env.GetCloudFormationStackName(), cfg.Logger.Dt)
	}
}

func deployStack(ctx *ecso.CommandContext, env *ecso.Environment, dryRun bool) error {
	var (
		cfg = ctx.Config

		stackName = env.GetCloudFormationStackName()
		template  = env.GetCloudFormationTemplateFile()
		prefix    = env.GetCloudFormationBucketPrefix()
		bucket    = env.CloudFormationBucket
		params    = env.CloudFormationParameters
		tags      = env.CloudFormationTags
	)

	cfnService, err := cfg.CloudFormationService(env.Region)

	if err != nil {
		return err
	}

	packagedTemplate, err := cfnService.Package(template, bucket, prefix)

	if err != nil {
		return err
	}

	cfg.Logger.Printf("\n")
	cfg.Logger.Infof("Deploying infrastructure stack '%s'", stackName)

	result, err := cfnService.Deploy(packagedTemplate, stackName, params, tags, dryRun)

	if err != nil {
		return err
	}

	changeSet, err := cfnService.GetChangeSet(result.ChangeSetID)

	if err != nil {
		return err
	}

	if dryRun {
		cfg.Logger.BannerGreen("The following changes would be made to the environment:")
	} else {
		cfg.Logger.BannerGreen("The following changes were made to the environment:")
	}

	fmt.Printf("%#v\n", changeSet)

	return nil
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
