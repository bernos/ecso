package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewEnvironmentPsCommand(environmentName string, environmentAPI api.EnvironmentAPI) ecso.Command {
	return &envPsCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type envPsCommand struct {
	*EnvironmentCommand
}

func (cmd *envPsCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	containers, err := cmd.environmentAPI.GetECSContainers(cmd.Environment(ctx))
	if err != nil {
		return err
	}

	_, err = containers.WriteTo(w)

	return err
}
