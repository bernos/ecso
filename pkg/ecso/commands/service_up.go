package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
	"gopkg.in/urfave/cli.v1"
)

const ServiceUpEnvironmentOption = "environment"

func NewServiceUpCommand(name string) ecso.Command {
	return &serviceUpCommand{
		name: name,
	}
}

type serviceUpCommand struct {
	name        string
	environment string
}

func (cmd *serviceUpCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceUpEnvironmentOption)
	return nil
}

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		cfg     = ctx.Config
		log     = cfg.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.environment]
		service = project.Services[cmd.name]
		ecsoAPI = api.New(cfg)
	)

	ui.BannerBlue(
		log,
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	if err := ecsoAPI.ServiceUp(project, env, service); err != nil {
		return err
	}

	description, err := ecsoAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(log, description)

	ui.BannerGreen(
		log,
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	return nil
}

func (cmd *serviceUpCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceUpCommand) Validate(ctx *ecso.CommandContext) error {
	err := util.AnyError(
		ui.ValidateRequired("Name")(cmd.name),
		ui.ValidateRequired("Environment")(cmd.environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[cmd.name]; !ok {
		return fmt.Errorf("Service '%s' not found", cmd.name)
	}

	if _, ok := ctx.Project.Environments[cmd.environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", cmd.environment)
	}

	return nil
}

func getTemplateDir(serviceName string) (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return wd, err
	}

	return filepath.Join(wd, ".ecso", "helpers", serviceName), nil
}
