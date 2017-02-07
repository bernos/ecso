package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/templates"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type EnvironmentUpOptions struct {
	EnvironmentName string
	DryRun          bool
}

func NewEnvironmentUpCommand(environmentName string, options ...func(*EnvironmentUpOptions)) ecso.Command {
	o := &EnvironmentUpOptions{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &envUpCommand{
		options: o,
	}
}

type envUpCommand struct {
	options *EnvironmentUpOptions
}

func (cmd *envUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		cfg     = ctx.Config
		env     = project.Environments[cmd.options.EnvironmentName]
		ecsoAPI = api.New(ctx.Config)
	)

	cfg.Logger.BannerBlue("Bringing up environment '%s'", env.Name)

	if cmd.options.DryRun {
		cfg.Logger.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := cmd.ensureTemplates(project, env, cfg.Logger); err != nil {
		return err
	}

	if err := ecsoAPI.EnvironmentUp(project, env, cmd.options.DryRun); err != nil {
		return err
	}

	if cmd.options.DryRun {
		cfg.Logger.BannerGreen("Review the above changes and re-run the command without the --dry-run option to apply them")

		return nil
	}

	cfg.Logger.BannerGreen("Environment '%s' is up and running", env.Name)

	return cmd.logEnvironmentDetails(ctx, env)
}

func (cmd *envUpCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if opt.EnvironmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(opt.EnvironmentName) {
		return fmt.Errorf("No environment named '%s' was found", opt.EnvironmentName)
	}

	return nil
}

func (cmd *envUpCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *envUpCommand) logEnvironmentDetails(ctx *ecso.CommandContext, env *ecso.Environment) error {
	var (
		log        = ctx.Config.Logger
		reg        = ctx.Config.MustGetAWSClientRegistry(env.Region)
		cfn        = reg.CloudFormationService(log.PrefixPrintf("  "))
		stack      = env.GetCloudFormationStackName()
		cfnConsole = util.CloudFormationConsoleURL(stack, env.Region)
		ecsConsole = util.ClusterConsoleURL(env.GetClusterName(), env.Region)
	)

	log.Dt("Cloud Formation console", cfnConsole)
	log.Dt("ECS Console", ecsConsole)

	return cfn.LogStackOutputs(stack, log.Dt)
}

func (cmd *envUpCommand) ensureTemplates(project *ecso.Project, env *ecso.Environment, logger ecso.Logger) error {
	dst := env.GetCloudFormationTemplateDir()

	exists, err := util.DirExists(dst)

	if err != nil || exists {
		return err
	}

	return templates.WriteEnvironmentFiles(project, env, nil)
}
