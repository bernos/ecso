package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
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

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	description, err := cmd.serviceAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(l, description)

	return nil
}
