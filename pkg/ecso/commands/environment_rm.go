package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	EnvironmentRmForceOption = "force"
)

func NewEnvironmentRmCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &environmentRmCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type environmentRmCommand struct {
	*EnvironmentCommand
}

func (cmd *environmentRmCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	fmt.Fprint(w, ui.BlueBannerf("Removing '%s' environment", env.Name))

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	delete(project.Environments, env.Name)

	if err := project.Save(); err != nil {
		return err
	}

	fmt.Fprint(w, ui.GreenBannerf("Successfully removed '%s' environment", env.Name))

	return nil
}

func (cmd *environmentRmCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.EnvironmentCommand.Validate(ctx); err != nil {
		return err
	}

	force := ctx.Options.Bool(EnvironmentRmForceOption)
	if !force {
		return ecso.NewOptionRequiredError(EnvironmentRmForceOption)
	}

	return nil
}
