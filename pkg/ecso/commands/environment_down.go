package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type EnvironmentDownOptions struct {
	Name string
}

func NewEnvironmentDownCommand(name string, options ...func(*EnvironmentDownOptions)) ecso.Command {
	o := &EnvironmentDownOptions{
		Name: name,
	}

	for _, option := range options {
		option(o)
	}

	return &environmentDownCommand{
		options: o,
	}
}

type environmentDownCommand struct {
	options *EnvironmentDownOptions
}

func (cmd *environmentDownCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
		env     = ctx.Project.Environments[cmd.options.Name]
		ecsoAPI = api.New(ctx.Config)
	)

	log.BannerBlue("Stopping '%s' environment", env.Name)

	if err := ecsoAPI.EnvironmentDown(project, env); err != nil {
		return err
	}

	log.BannerGreen("Successfully stopped '%s' environment", env.Name)

	return nil
}

func (cmd *environmentDownCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *environmentDownCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if ctx.Project.Environments[opt.Name] == nil {
		return fmt.Errorf("Environment '%s' not found", opt.Name)
	}

	return nil
}
