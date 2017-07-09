package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceVersionsCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceVersionsCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceVersionsCommand struct {
	*ServiceCommand
}

func (cmd *serviceVersionsCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	versions, err := cmd.serviceAPI.GetAvailableVersions(ctx.Project, env, service)
	if err != nil {
		return err
	}

	ui.PrintTable(w, versions)

	return nil
}
