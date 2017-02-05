package servicedown

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type Options struct {
	Name        string
	Environment string
}

func New(name, environment string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &command{
		options: o,
	}
}

type command struct {
	options *Options
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	var (
		ecsoAPI = api.New(ctx.Config)
		service = ctx.Project.Services[cmd.options.Name]
		env     = ctx.Project.Environments[cmd.options.Environment]
		log     = ctx.Config.Logger
	)

	log.BannerBlue(
		"Terminating the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	if err := ecsoAPI.ServiceDown(ctx.Project, env, service); err != nil {
		return err
	}

	log.BannerGreen(
		"Successfully terminated the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if opt.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if opt.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(opt.Name) {
		return fmt.Errorf("No service named '%s' was found", opt.Name)
	}

	if !ctx.Project.HasEnvironment(opt.Environment) {
		return fmt.Errorf("No environment named '%s' was found", opt.Environment)
	}

	return nil
}
