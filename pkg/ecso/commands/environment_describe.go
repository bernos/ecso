package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewEnvironmentDescribeCommand(environmentName string, environmentAPI api.EnvironmentAPI, log ecso.Logger) ecso.Command {
	return &environmentDescribeCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
		},
		environmentAPI: environmentAPI,
		log:            log,
	}
}

type environmentDescribeCommand struct {
	*EnvironmentCommand

	log            ecso.Logger
	environmentAPI api.EnvironmentAPI
}

func (cmd *environmentDescribeCommand) Execute(ctx *ecso.CommandContext) error {
	env := ctx.Project.Environments[cmd.environmentName]

	description, err := cmd.environmentAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

	ui.PrintEnvironmentDescription(cmd.log, description)

	return nil
}
