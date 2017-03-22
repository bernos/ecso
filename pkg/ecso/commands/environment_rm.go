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

func NewEnvironmentRmCommand(environmentName string, environmentAPI api.EnvironmentAPI, log ecso.Logger) ecso.Command {
	return &environmentRmCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
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
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environmentName]
	)

	ui.BannerBlue(cmd.log, "Removing '%s' environment", env.Name)

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	delete(project.Environments, cmd.environmentName)

	if err := project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(cmd.log, "Successfully removed '%s' environment", env.Name)

	return nil
}
