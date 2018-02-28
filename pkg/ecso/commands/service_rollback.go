package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewServiceRollbackCommand(name string, environmentName string, version string, serviceAPI api.ServiceAPI) *ServiceRollbackCommand {
	return &ServiceRollbackCommand{
		ServiceCommand: &ServiceCommand{
			name:            name,
			environmentName: environmentName,
			serviceAPI:      serviceAPI,
		},
		version: version,
	}
}

type ServiceRollbackCommand struct {
	*ServiceCommand
	version string
}

func (cmd *ServiceRollbackCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.ServiceCommand.Validate(ctx); err != nil {
		return err
	}

	if cmd.version == "" {
		return fmt.Errorf("Version is required")
	}

	return nil
}

func (cmd *ServiceRollbackCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
	)

	fmt.Fprintf(blue, "Rolling back service '%s' to version '%s' in the '%s' environment", service.Name, cmd.version, env.Name)

	description, err := cmd.serviceAPI.ServiceRollback(project, env, service, cmd.version, w)
	if err != nil {
		return err
	}

	description.WriteTo(w)

	fmt.Fprintf(green, "Rolled back service '%s' to version '%s' in the '%s' environment", service.Name, cmd.version, env.Name)

	return nil
}
