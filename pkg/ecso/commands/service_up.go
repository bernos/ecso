package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceUpOptions struct {
	Name        string
	Environment string
}

func NewServiceUpCommand(name, environment string, options ...func(*ServiceUpOptions)) ecso.Command {
	o := &ServiceUpOptions{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &serviceUpCommand{
		options: o,
	}
}

type serviceUpCommand struct {
	options *ServiceUpOptions
}

func (cmd *serviceUpCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		cfg     = ctx.Config
		log     = cfg.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.options.Environment]
		service = project.Services[cmd.options.Name]
		ecsoAPI = api.New(cfg)
	)

	log.BannerBlue(
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	if err := ecsoAPI.ServiceUp(project, env, service); err != nil {
		return err
	}

	log.BannerGreen(
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		env.Name)

	description, err := ecsoAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(description, log)

	return nil
}

func (cmd *serviceUpCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceUpCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	err := util.AnyError(
		ui.ValidateRequired("Name")(opt.Name),
		ui.ValidateRequired("Environment")(opt.Environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[opt.Name]; !ok {
		return fmt.Errorf("Service '%s' not found", opt.Name)
	}

	if _, ok := ctx.Project.Environments[opt.Environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
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
