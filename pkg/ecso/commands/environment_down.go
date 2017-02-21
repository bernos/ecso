package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentDownCommand(environmentName string) ecso.Command {
	return &environmentDownCommand{
		environmentName: environmentName,
	}
}

type environmentDownCommand struct {
	environmentName string
}

func (cmd *environmentDownCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
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

func (cmd *environmentDownCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *environmentDownCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if ctx.Project.Environments[cmd.environmentName] == nil {
		return fmt.Errorf("Environment '%s' not found", cmd.environmentName)
	}

	return nil
}
