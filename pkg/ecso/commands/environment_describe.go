package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewEnvironmentDescribeCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
	return &environmentDescribeCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
		},
	}
}

type environmentDescribeCommand struct {
	*EnvironmentCommand
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
