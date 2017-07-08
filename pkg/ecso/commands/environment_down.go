package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
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

func (cmd *environmentDownCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	ui.BannerBlue(l, "Stopping '%s' environment", env.Name)

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	ui.BannerGreen(l, "Successfully stopped '%s' environment", env.Name)

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
