package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceDescribeCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceDecribeCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceDecribeCommand struct {
	*ServiceCommand
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	_, err := cmd.serviceAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	// ui.PrintServiceDescription(l, description)

	return nil
}
