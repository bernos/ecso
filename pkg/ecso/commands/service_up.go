package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceUpCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceUpCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceUpCommand struct {
	*ServiceCommand
}

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	fmt.Fprint(w, ui.BlueBannerf("Deploying service '%s' to the '%s' environment", service.Name, env.Name))

	_, err := cmd.serviceAPI.ServiceUp(project, env, service)

	if err != nil {
		return err
	}

	// ui.PrintServiceDescription(l, description)

	fmt.Fprint(w, ui.GreenBannerf("Deployed service '%s' to the '%s' environment", service.Name, env.Name))

	return nil
}
