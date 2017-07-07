package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceVersionsCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &serviceVersionsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceVersionsCommand struct {
	*ServiceCommand
}

func (cmd *serviceVersionsCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	versions, err := cmd.serviceAPI.GetAvailableVersions(ctx.Project, env, service)
	if err != nil {
		return err
	}

	ui.PrintTable(cmd.log, versions)

	return nil
}
