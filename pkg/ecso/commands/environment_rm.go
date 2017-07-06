package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	EnvironmentRmForceOption = "force"
)

func NewEnvironmentRmCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
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

func (cmd *environmentRmCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	ui.BannerBlue(cmd.log, "Removing '%s' environment", env.Name)

	if err := cmd.environmentAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	delete(project.Environments, env.Name)

	if err := project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(cmd.log, "Successfully removed '%s' environment", env.Name)

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
