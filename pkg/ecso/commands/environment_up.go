package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/resources"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

const (
	EnvironmentUpDryRunOption = "dry-run"
	EnvironmentUpForceOption  = "force"
)

func NewEnvironmentUpCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &envUpCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type envUpCommand struct {
	*EnvironmentCommand

	dryRun bool
	force  bool
}

func (cmd *envUpCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	cmd.dryRun = ctx.Options.Bool(EnvironmentUpDryRunOption)
	cmd.force = ctx.Options.Bool(EnvironmentUpForceOption)

	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
		info    = ui.NewInfoWriter(w)
	)

	fmt.Fprintf(blue, "Bringing up environment '%s'", env.Name)

	if cmd.dryRun {
		fmt.Fprintf(info, "THIS IS A DRY RUN - no changes to the environment will be made.")
	}

	if err := cmd.ensureTemplates(ctx, project, env); err != nil {
		return err
	}

	if err := cmd.ensureResources(ctx, project, env); err != nil {
		return err
	}

	if err := cmd.environmentAPI.EnvironmentUp(project, env, cmd.dryRun); err != nil {
		return err
	}

	if cmd.dryRun {
		fmt.Fprintf(green, "Review the above changes and re-run the command without the --dry-run option to apply them")
		return nil
	}

	fmt.Fprintf(green, "Environment '%s' is up and running", env.Name)

	_, err := cmd.environmentAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

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
}

func (cmd *envUpCommand) ensureResources(ctx *ecso.CommandContext, project *ecso.Project, env *ecso.Environment) error {
	dst := env.GetResourceDir()

	exists, err := util.DirExists(dst)
	if err != nil || exists {
		return err
	}

	return resources.EnvironmentResources.WriteTo(dst, nil)
}
