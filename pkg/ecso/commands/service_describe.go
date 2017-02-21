package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
	"gopkg.in/urfave/cli.v1"
)

const ServiceDescribeEnvironmentOption = "environment"

func NewServiceDescribeCommand(name string) ecso.Command {
	return &serviceDecribeCommand{
		name: name,
	}
}

type serviceDecribeCommand struct {
	name        string
	environment string
}

func (cmd *serviceDecribeCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceDescribeEnvironmentOption)
	return nil
}

func (cmd *serviceDecribeCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		env     = ctx.Project.Environments[cmd.environment]
		service = ctx.Project.Services[cmd.name]
		log     = ctx.Config.Logger()
		ecsoAPI = api.New(ctx.Config)
	)

	description, err := ecsoAPI.DescribeService(env, service)

	if err != nil {
		return err
	}

	ui.PrintServiceDescription(log, description)

	return nil
}

func (cmd *serviceDecribeCommand) Validate(ctx *ecso.CommandContext) error {
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

func (cmd *serviceDecribeCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
