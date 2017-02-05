package rm

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type Options struct {
	Name string
}

func New(name string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name: name,
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
		log     = ctx.Config.Logger
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.options.Name]
		ecsoAPI = api.New(ctx.Config)
	)

	log.BannerBlue("Removing '%s' environment", env.Name)

	if err := ecsoAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	delete(project.Environments, cmd.options.Name)

	if err := project.Save(); err != nil {
		return err
	}

	log.BannerGreen("Successfully removed '%s' environment", env.Name)

	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if ctx.Project.Environments[opt.Name] == nil {
		return fmt.Errorf("Environment '%s' not found", opt.Name)
	}

	return nil
}
