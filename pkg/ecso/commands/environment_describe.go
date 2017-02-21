package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvironmentDescribeCommand(environmentName string) ecso.Command {
	return &environmentDescribeCommand{
		environmentName: environmentName,
	}
}

type environmentDescribeCommand struct {
	environmentName string
}

func (cmd *environmentDescribeCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
}

func (cmd *environmentDescribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.environmentName]
		log     = ctx.Config.Logger()
		ecsoAPI = api.New(ctx.Config)
	)

	description, err := ecsoAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

	ui.PrintEnvironmentDescription(log, description)

	return nil
}

func (cmd *environmentDescribeCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(cmd.environmentName) {
		return fmt.Errorf("No environment named '%s' was found", cmd.environmentName)
	}

	return nil
}

func (cmd *environmentDescribeCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
