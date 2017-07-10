package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
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

func (cmd *environmentDescribeCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	description, err := cmd.environmentAPI.DescribeEnvironment(cmd.Environment(ctx))
	if err != nil {
		return err
	}

	description.WriteTo(w)

	return nil
}
