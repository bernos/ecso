package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
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

func (cmd *environmentDownCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environmentName]
		ecsoAPI = api.New(ctx.Config)
	)

	ui.BannerBlue(log, "Stopping '%s' environment", env.Name)

	if err := ecsoAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	ui.BannerGreen(log, "Successfully stopped '%s' environment", env.Name)

	return nil
}
