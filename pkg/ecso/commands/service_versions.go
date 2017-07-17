package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceVersionsCommand(name string, environmentName string, serviceAPI api.ServiceAPI) *ServiceVersionsCommand {
	return &ServiceVersionsCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type ServiceVersionsCommand struct {
	*ServiceCommand
}

func (cmd *ServiceVersionsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	versions, err := cmd.serviceAPI.GetAvailableVersions(ctx.Project, env, service)
	if err != nil {
		return err
	}

	_, err = versions.WriteTo(w)

	return err
}
