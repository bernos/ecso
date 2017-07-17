package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServicePsCommand(name string, environmentName string, serviceAPI api.ServiceAPI) ecso.Command {
	return &servicePsCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type servicePsCommand struct {
	*ServiceCommand
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	containers, err := cmd.serviceAPI.GetECSContainers(ctx.Project, env, service)
	if err != nil {
		return err
	}

	_, err = containers.WriteTo(w)

	return err
}
