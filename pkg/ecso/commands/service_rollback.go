package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	ServiceRollbackVersionOption = "version"
)

func NewServiceRollbackCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &serviceRollbackCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
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

func (cmd *serviceRollbackCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
		version = ctx.Options.String(ServiceRollbackVersionOption)
	)

	ui.BannerBlue(
		cmd.log,
		"Rolling back service '%s' to version '%s' in the '%s' environment",
		service.Name,
		version,
		env.Name)

	description, err := cmd.serviceAPI.ServiceRollback(project, env, service, version)
	if err != nil {
		return err
	}

	ui.PrintServiceDescription(cmd.log, description)

	ui.BannerGreen(
		cmd.log,
		"Rolled back service '%s' to version '%s' in the '%s' environment",
		service.Name,
		version,
		env.Name)

	return nil
}
