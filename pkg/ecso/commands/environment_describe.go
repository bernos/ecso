package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewEnvironmentDescribeCommand(environmentName string) ecso.Command {
	return &environmentDescribeCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
		},
	}
}

type environmentDescribeCommand struct {
	*EnvironmentCommand
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
