package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentDownForceOption = "force"
)

func NewEnvironmentDownCommand(environmentName string) ecso.Command {
	return &environmentDownCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
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
		log     = ctx.Config.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environmentName]
		ecsoAPI = api.NewEnvironmentAPI(ctx.Config)
	)

	ui.BannerBlue(log, "Stopping '%s' environment", env.Name)

	if err := ecsoAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	ui.BannerGreen(log, "Successfully stopped '%s' environment", env.Name)

	return nil
}
