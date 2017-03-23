package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceDownForceOption = "force"
)

func NewServiceDownCommand(name string, serviceAPI api.ServiceAPI, log log.Logger) ecso.Command {
	return &serviceDownCommand{
		ServiceCommand: &ServiceCommand{
			name:       name,
			serviceAPI: serviceAPI,
			log:        log,
		},
	}
}

type serviceDownCommand struct {
	*ServiceCommand
}

func (cmd *serviceDownCommand) UnmarshalCliContext(ctx *cli.Context) error {
	if err := cmd.ServiceCommand.UnmarshalCliContext(ctx); err != nil {
		return err
	}

	force := ctx.Bool(ServiceDownForceOption)

	if !force {
		return ecso.NewOptionRequiredError(ServiceDownForceOption)
	}

	return nil
}

func (cmd *serviceDownCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = cmd.Environment(ctx)
		service = cmd.Service(ctx)
	)

	ui.BannerBlue(
		cmd.log,
		"Terminating the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	if err := cmd.serviceAPI.ServiceDown(ctx.Project, env, service); err != nil {
		return err
	}

	ui.BannerGreen(
		cmd.log,
		"Successfully terminated the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	return nil
}
