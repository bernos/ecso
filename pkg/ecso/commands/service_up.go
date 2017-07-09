package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
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

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	ui.BannerBlue(
		l,
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	_, err := cmd.serviceAPI.ServiceUp(project, env, service)

	if err != nil {
		return err
	}

	// ui.PrintServiceDescription(l, description)

	ui.BannerGreen(
		l,
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	return nil
}
