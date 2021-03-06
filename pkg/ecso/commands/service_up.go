package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceUpCommand(name string, environmentName string, serviceAPI api.ServiceAPI) *ServiceUpCommand {
	return &ServiceUpCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type ServiceUpCommand struct {
	*ServiceCommand
}

func (cmd *ServiceUpCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
	)

	fmt.Fprintf(blue, "Deploying service '%s' to the '%s' environment", service.Name, env.Name)

	description, err := cmd.serviceAPI.ServiceUp(project, env, service, w)

	if err != nil {
		return err
	}

	description.WriteTo(w)

	fmt.Fprintf(green, "Deployed service '%s' to the '%s' environment", service.Name, env.Name)

	return nil
}
