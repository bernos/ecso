package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
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
	version string
}

func (cmd *serviceRollbackCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := cmd.ServiceCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	cmd.version = ctx.String(ServiceRollbackVersionOption)

	return nil
}

func (cmd *serviceRollbackCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.ServiceCommand.Validate(ctx); err != nil {
		return err
	}

	if cmd.version == "" {
		return fmt.Errorf("Version is required")
	}

	return nil
}

func (cmd *serviceRollbackCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	ui.BannerBlue(
		cmd.log,
		"Rolling back service '%s' to version '%s' in the '%s' environment",
		service.Name,
		cmd.version,
		env.Name)

	description, err := cmd.serviceAPI.ServiceRollback(project, env, service, cmd.version)
	if err != nil {
		return err
	}

	ui.PrintServiceDescription(cmd.log, description)

	ui.BannerGreen(
		cmd.log,
		"Rolled back service '%s' to version '%s' in the '%s' environment",
		service.Name,
		cmd.version,
		env.Name)

	return nil
}
