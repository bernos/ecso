package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

func NewEnvironmentDescribeCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &environmentDescribeCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type environmentDescribeCommand struct {
	*EnvironmentCommand
}

func (cmd *environmentDescribeCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	_, err := cmd.environmentAPI.DescribeEnvironment(cmd.Environment(ctx))

	if err != nil {
		return err
	}

	// ui.PrintEnvironmentDescription(l, description)

	return nil
}
