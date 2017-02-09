package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type EnvironmentDescribeOptions struct {
	EnvironmentName string
}

func NewEnvironmentDescribeCommand(environmentName string, options ...func(*EnvironmentDescribeOptions)) ecso.Command {
	o := &EnvironmentDescribeOptions{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &environmentDescribeCommand{
		options: o,
	}
}

type environmentDescribeCommand struct {
	options *EnvironmentDescribeOptions
}

func (cmd *environmentDescribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.options.EnvironmentName]
		log     = ctx.Config.Logger()
		ecsoAPI = api.New(ctx.Config)
	)

	description, err := ecsoAPI.DescribeEnvironment(env)

	if err != nil {
		return err
	}

	log.BannerGreen("Details of the '%s' environment", env.Name)
	log.Dl(description)
	log.Printf("\n")

	return nil
}

func (cmd *environmentDescribeCommand) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if opt.EnvironmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(opt.EnvironmentName) {
		return fmt.Errorf("No environment named '%s' was found", opt.EnvironmentName)
	}

	return nil
}

func (cmd *environmentDescribeCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
