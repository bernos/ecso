package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentDownForceOption = "force"
)

func NewEnvironmentDownCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
	return &environmentDownCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
		},
	}
}

type environmentDownCommand struct {
	*EnvironmentCommand
}

func (cmd *environmentDownCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := cmd.EnvironmentCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	force := ctx.Bool(EnvironmentDownForceOption)

	if !force {
		return ecso.NewOptionRequiredError(EnvironmentDownForceOption)
	}

	return nil
}

func (cmd *environmentDownCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	ui.BannerBlue(cmd.log, "Stopping '%s' environment", env.Name)

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	ui.BannerGreen(cmd.log, "Successfully stopped '%s' environment", env.Name)

	return nil
}
