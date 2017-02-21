package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceDownEnvironmentOption = "environment"
)

func NewServiceDownCommand(name string) ecso.Command {
	return &serviceDownCommand{
		name: name,
	}
}

type serviceDownCommand struct {
	name        string
	environment string
}

func (cmd *serviceDownCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceDownEnvironmentOption)
	return nil
}

func (cmd *serviceDownCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		ecsoAPI = api.New(ctx.Config)
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

func (cmd *serviceDownCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceDownCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.name == "" {
		return fmt.Errorf("Name is required")
	}

	if cmd.environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(cmd.name) {
		return fmt.Errorf("No service named '%s' was found", cmd.name)
	}

	if !ctx.Project.HasEnvironment(cmd.environment) {
		return fmt.Errorf("No environment named '%s' was found", cmd.environment)
	}

	return nil
}
