package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceDownCommand(name string, environmentName string, serviceAPI api.ServiceAPI) *ServiceDownCommand {
	return &ServiceDownCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
	}
}

type ServiceDownCommand struct {
	*ServiceCommand
	force bool
}

func (cmd *ServiceDownCommand) WithForce(force bool) *ServiceDownCommand {
	cmd.force = force
	return cmd
}

func (cmd *ServiceDownCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
	)

	fmt.Fprintf(blue, "Terminating the '%s' service in the '%s' environment", service.Name, env.Name)

	if err := cmd.serviceAPI.ServiceDown(ctx.Project, env, service, w); err != nil {
		return err
	}

	fmt.Fprintf(green, "Successfully terminated the '%s' service in the '%s' environment", service.Name, env.Name)

	return nil
}

func (cmd *ServiceDownCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.ServiceCommand.Validate(ctx); err != nil {
		return err
	}

	if !cmd.force {
		return ecso.NewOptionRequiredError("force")
	}

	return nil
}
