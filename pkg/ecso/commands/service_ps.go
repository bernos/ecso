package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServicePsCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &servicePsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type servicePsCommand struct {
	*ServiceCommand
}

func (cmd *servicePsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	containers, err := cmd.serviceAPI.GetECSContainers(ctx.Project, env, service)

	if err != nil {
		return err
	}

	ui.PrintTable(cmd.log, containers)

	return nil
}
