package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentRmCommand(environmentName string) ecso.Command {
	return &environmentRmCommand{
		environmentName: environmentName,
	}
}

type environmentRmCommand struct {
	environmentName string
}

func (cmd *environmentRmCommand) UnmarshalCliContext(ctx *cli.Context) error {
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

func (cmd *environmentRmCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *environmentRmCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if ctx.Project.Environments[cmd.environmentName] == nil {
		return fmt.Errorf("Environment '%s' not found", cmd.environmentName)
	}

	return nil
}
