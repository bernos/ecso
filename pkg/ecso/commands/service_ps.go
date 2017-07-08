package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServicePsCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &servicePsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type servicePsCommand struct {
	*ServiceCommand
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	containers, err := cmd.serviceAPI.GetECSContainers(ctx.Project, env, service)

	if err != nil {
		return err
	}

	ui.PrintTable(l, containers)

	return nil
}
