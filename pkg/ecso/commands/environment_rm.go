package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewEnvironmentRmCommand(environmentName string, environmentAPI api.EnvironmentAPI) *EnvironmentRmCommand {
	return &EnvironmentRmCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type EnvironmentRmCommand struct {
	*EnvironmentCommand
	force bool
}

func (cmd *EnvironmentRmCommand) WithForce(force bool) *EnvironmentRmCommand {
	cmd.force = force
	return cmd
}

func (cmd *EnvironmentRmCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
	)

	fmt.Fprintf(blue, "Removing '%s' environment", env.Name)

	if err := cmd.environmentAPI.EnvironmentDown(project, env, w); err != nil {
		return err
	}

	delete(project.Environments, env.Name)

	if err := project.Save(); err != nil {
		return err
	}

	fmt.Fprintf(green, "Successfully removed '%s' environment", env.Name)

	return nil
}

func (cmd *EnvironmentRmCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.EnvironmentCommand.Validate(ctx); err != nil {
		return err
	}

	if !cmd.force {
		return ecso.NewOptionRequiredError("force")
	}

	return nil
}
