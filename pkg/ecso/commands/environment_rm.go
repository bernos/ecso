package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	EnvironmentRmForceOption = "force"
)

func NewEnvironmentRmCommand(environmentName string) ecso.Command {
	return &environmentRmCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
		},
	}
}

type environmentRmCommand struct {
	*EnvironmentCommand
}

func (cmd *environmentRmCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := cmd.EnvironmentCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	force := ctx.Bool(EnvironmentRmForceOption)

	if !force {
		return ecso.NewOptionRequiredError(EnvironmentRmForceOption)
	}

	return nil
}

func (cmd *environmentRmCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environmentName]
		ecsoAPI = api.New(ctx.Config)
	)

	ui.BannerBlue(log, "Removing '%s' environment", env.Name)

	if err := ecsoAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	delete(project.Environments, cmd.environmentName)

	if err := project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(log, "Successfully removed '%s' environment", env.Name)

	return nil
}
