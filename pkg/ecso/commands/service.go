package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceEnvironmentOption = "environment"
)

type ServiceCommand struct {
	name        string
	environment string
	serviceAPI  api.ServiceAPI
	log         ecso.Logger
}

func (cmd *ServiceCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.environment = ctx.String(ServiceEnvironmentOption)
	return nil
}

func (cmd *ServiceCommand) Validate(ctx *ecso.CommandContext) error {
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

func (cmd *ServiceCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}
