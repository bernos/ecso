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

func NewEnvironmentUpCommand(environmentName string, environmentAPI api.EnvironmentAPI) *EnvironmentUpCommand {
	return &EnvironmentUpCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type EnvironmentUpCommand struct {
	*EnvironmentCommand

	dryRun bool
	force  bool
}

func (cmd *EnvironmentUpCommand) WithDryRun(dryRun bool) *EnvironmentUpCommand {
	cmd.dryRun = dryRun
	return cmd
}

func (cmd *EnvironmentUpCommand) WithForce(force bool) *EnvironmentUpCommand {
	cmd.force = force
	return cmd
}

func (cmd *EnvironmentUpCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
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

	if err := cmd.ensureEnvironmentFiles(ctx, project, env); err != nil {
		return err
	}

	if err := cmd.environmentAPI.EnvironmentUp(project, env, cmd.dryRun, w); err != nil {
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

func (cmd *EnvironmentUpCommand) ensureEnvironmentFiles(ctx *ecso.CommandContext, project *ecso.Project, env *ecso.Environment) error {
	exists, err := util.DirExists(env.GetCloudFormationTemplateDir())
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

	w := resources.NewFileSystemResourceWriter(project.Dir())
	return w.WriteResources(nil, resources.EnvironmentFiles...)
}
