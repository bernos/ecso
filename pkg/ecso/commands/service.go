package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
)

const (
	ServiceEnvironmentOption = "environment"
)

type ServiceCommand struct {
	name       string
	serviceAPI api.ServiceAPI
}

func (cmd *ServiceCommand) Environment(ctx *ecso.CommandContext) *ecso.Environment {
	environmentName := ctx.Options.String(ServiceEnvironmentOption)

	if environmentName == "" {
		return nil
	}

	return ctx.Project.Environments[environmentName]
}

func (cmd *ServiceCommand) Service(ctx *ecso.CommandContext) *ecso.Service {
	return ctx.Project.Services[cmd.name]
}

func (cmd *ServiceCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.name == "" {
		return fmt.Errorf("Name is required")
	}

	if ctx.Options.String(ServiceEnvironmentOption) == "" {
		return fmt.Errorf("Environment is required")
	}

	if cmd.Environment(ctx) == nil {
		return fmt.Errorf("No environment named '%s' was found", cmd.Environment(ctx).Name)
	}

	if cmd.Service(ctx) == nil {
		return fmt.Errorf("No service named '%s' was found", cmd.name)
	}

	return nil
}

func (cmd *ServiceCommand) Prompt(ctx *ecso.CommandContext, l log.Logger) error {
	return nil
}
