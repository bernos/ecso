package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceDescribeOptions struct {
	Name        string
	Environment string
}

func NewServiceDescribeCommand(name, environment string, options ...func(*ServiceDescribeOptions)) ecso.Command {
	o := &ServiceDescribeOptions{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &serviceDecribeCommand{
		options: o,
	}
}

type serviceDecribeCommand struct {
	options *ServiceDescribeOptions
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.options.Environment]
		service = ctx.Project.Services[cmd.options.Name]
		log     = ctx.Config.Logger()
		ecsoAPI = api.New(ctx.Config)
	)

	description, err := ecsoAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(description, log)

	return nil
}

func (cmd *serviceDecribeCommand) Validate(ctx *ecso.CommandContext) error {
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

func (cmd *serviceDecribeCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
