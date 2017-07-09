package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	EnvironmentDownForceOption = "force"
)

func NewEnvironmentDownCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &environmentDownCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type environmentDownCommand struct {
	*EnvironmentCommand
}

func (cmd *environmentDownCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	fmt.Fprint(w, ui.BlueBannerf("Stopping '%s' environment", env.Name))

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	fmt.Fprint(w, ui.GreenBannerf("Successfully stopped '%s' environment", env.Name))

	return nil
}

func (cmd *environmentDownCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.EnvironmentCommand.Validate(ctx); err != nil {
		return err
	}

	force := ctx.Options.Bool(EnvironmentDownForceOption)
	if !force {
		return ecso.NewOptionRequiredError(EnvironmentDownForceOption)
	}

	return nil
}
