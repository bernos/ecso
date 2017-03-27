package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewEnvironmentPsCommand(environmentName string, environmentAPI api.EnvironmentAPI, log log.Logger) ecso.Command {
	return &envPsCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
			log:             log,
		},
	}
}

type envPsCommand struct {
	*EnvironmentCommand
}

func (cmd *envPsCommand) Execute(ctx *ecso.CommandContext) error {
	containers, err := cmd.environmentAPI.GetECSContainers(cmd.Environment(ctx))

	if err != nil {
		return err
	}

	ui.PrintTable(cmd.log, containers)

	return nil
}
