package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceDescribeCommand(name string, environmentName string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceDecribeCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type serviceDecribeCommand struct {
	*ServiceCommand
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	description, err := cmd.serviceAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	description.WriteTo(w)

	return nil
}
