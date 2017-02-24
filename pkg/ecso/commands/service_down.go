package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceDownForceOption = "force"
)

func NewServiceDownCommand(name string) ecso.Command {
	return &serviceDownCommand{
		ServiceCommand: &ServiceCommand{
			name: name,
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
		ecsoAPI = api.NewServiceAPI(ctx.Config)
		service = ctx.Project.Services[cmd.name]
		env     = ctx.Project.Environments[cmd.environment]
		log     = ctx.Config.Logger()
	)

	ui.BannerBlue(
		log,
		"Terminating the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	if err := ecsoAPI.ServiceDown(ctx.Project, env, service); err != nil {
		return err
	}

	ui.BannerGreen(
		log,
		"Successfully terminated the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	return nil
}
