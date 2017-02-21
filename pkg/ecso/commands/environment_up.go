package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/templates"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"

	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentUpDryRunOption = "dry-run"
)

func NewEnvironmentUpCommand(environmentName string) ecso.Command {
	return &envUpCommand{
		environmentName: environmentName,
	}
}

type envUpCommand struct {
	environmentName string
	dryRun          bool
}

func (cmd *envUpCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.dryRun = ctx.Bool(EnvironmentUpDryRunOption)

	return nil
}

func (cmd *envUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		cfg     = ctx.Config
		log     = cfg.Logger()
		env     = project.Environments[cmd.environmentName]
		ecsoAPI = api.New(ctx.Config)
	)

	ui.BannerBlue(log, "Bringing up environment '%s'", env.Name)

	if cmd.dryRun {
		log.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := cmd.ensureTemplates(project, env, log); err != nil {
		return err
	}

	if err := ecsoAPI.EnvironmentUp(project, env, cmd.dryRun); err != nil {
		return err
	}

	if cmd.dryRun {
		ui.BannerGreen(log, "Review the above changes and re-run the command without the --dry-run option to apply them")

		return nil
	}

	ui.BannerGreen(log, "Environment '%s' is up and running", env.Name)

	description, err := ecsoAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

	ui.PrintEnvironmentDescription(log, description)

	return nil
}

func (cmd *envUpCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(cmd.environmentName) {
		return fmt.Errorf("No environment named '%s' was found", cmd.environmentName)
	}

	return nil
}

func (cmd *envUpCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *envUpCommand) ensureTemplates(project *ecso.Project, env *ecso.Environment, logger ecso.Logger) error {
	dst := env.GetCloudFormationTemplateDir()

	exists, err := util.DirExists(dst)

	if err != nil || exists {
		return err
	}

	return templates.WriteEnvironmentFiles(project, env, nil)
}
