package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	ServiceRollbackVersionOption = "version"
)

func NewServiceRollbackCommand(name string, serviceAPI api.ServiceAPI) ecso.Command {
	return &serviceRollbackCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
		},
	}
}

type serviceRollbackCommand struct {
	*ServiceCommand
}

func (cmd *serviceRollbackCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.ServiceCommand.Validate(ctx); err != nil {
		return err
	}

	if ctx.Options.String(ServiceRollbackVersionOption) == "" {
		return fmt.Errorf("Version is required")
	}

	return nil
}

func (cmd *serviceRollbackCommand) Execute(ctx *ecso.CommandContext, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		version = ctx.Options.String(ServiceRollbackVersionOption)
		blue    = ui.NewBannerWriter(w, ui.BlueBold)
		green   = ui.NewBannerWriter(w, ui.GreenBold)
	)

	fmt.Fprintf(blue, "Rolling back service '%s' to version '%s' in the '%s' environment", service.Name, version, env.Name)

	_, err := cmd.serviceAPI.ServiceRollback(project, env, service, version)
	if err != nil {
		return err
	}

	// ui.PrintServiceDescription(l, description)

	fmt.Fprintf(green, "Rolled back service '%s' to version '%s' in the '%s' environment", service.Name, version, env.Name)

	return nil
}
