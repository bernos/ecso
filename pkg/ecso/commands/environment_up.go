package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/resources"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"

	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentUpDryRunOption = "dry-run"
	EnvironmentUpForceOption  = "force"
)

func NewEnvironmentUpCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
	return &envUpCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
		},
	}
}

type envUpCommand struct {
	*EnvironmentCommand

	dryRun bool
	force  bool
}

func (cmd *envUpCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := cmd.EnvironmentCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	cmd.dryRun = ctx.Bool(EnvironmentUpDryRunOption)
	cmd.force = ctx.Bool(EnvironmentUpForceOption)

	return nil
}

func (cmd *envUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	ui.BannerBlue(cmd.log, "Bringing up environment '%s'", env.Name)

	if cmd.dryRun {
		cmd.log.Infof("THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := cmd.ensureTemplates(ctx, project, env); err != nil {
		return err
	}

	if err := cmd.environmentAPI.EnvironmentUp(project, env, cmd.dryRun); err != nil {
		return err
	}

	if cmd.dryRun {
		ui.BannerGreen(cmd.log, "Review the above changes and re-run the command without the --dry-run option to apply them")

		return nil
	}

	ui.BannerGreen(cmd.log, "Environment '%s' is up and running", env.Name)

	description, err := cmd.environmentAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

	ui.PrintEnvironmentDescription(cmd.log, description)

	return nil
}

func (cmd *envUpCommand) ensureTemplates(ctx *ecso.CommandContext, project *ecso.Project, env *ecso.Environment) error {
	dst := env.GetCloudFormationTemplateDir()

	exists, err := util.DirExists(dst)

	if err != nil || exists {
		return err
	}

	stackExists, err := cmd.environmentAPI.IsEnvironmentUp(env)

	if err != nil {
		return err
	}

	if stackExists && !cmd.force {
		return fmt.Errorf("This looks like the first time you've run `environment up` for the %s environment from this repository, however there is already a CloudFormation stack up and running. This could mean that someone has already created the %s environment for the %s project. If you really know what you are doing, you can rerun `environment up` with the `--force` flag.", env.Name, env.Name, project.Name)
	}

	return resources.EnvironmentCloudFormationTemplates.WriteTo(dst, nil)
	// return templates.WriteEnvironmentFiles(project, env, nil)
}
