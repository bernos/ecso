package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type EnvironmentRmOptions struct {
	Name string
}

func NewEnvironmentRmCommand(name string, options ...func(*EnvironmentRmOptions)) ecso.Command {
	o := &EnvironmentRmOptions{
		Name: name,
	}

	for _, option := range options {
		option(o)
	}

	return &environmentRmCommand{
		options: o,
	}
}

type environmentRmCommand struct {
	options *EnvironmentRmOptions
}

func (cmd *environmentRmCommand) Execute(ctx *ecso.CommandContext) error {
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

func (cmd *environmentRmCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *environmentRmCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if ctx.Project.Environments[opt.Name] == nil {
		return fmt.Errorf("Environment '%s' not found", opt.Name)
	}

	return nil
}
