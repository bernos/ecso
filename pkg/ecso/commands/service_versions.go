package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewServiceVersionsCommand(name string, environmentName string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceVersionsCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type serviceVersionsCommand struct {
	*ServiceCommand
}

func (cmd *serviceVersionsCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
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
