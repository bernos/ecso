package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceDescribeCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &serviceDecribeCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceDecribeCommand struct {
	*ServiceCommand
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.environment]
		service = ctx.Project.Services[cmd.name]
	)

	description, err := cmd.serviceAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(cmd.log, description)

	return nil
}
