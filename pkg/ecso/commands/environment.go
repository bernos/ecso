package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/log"
	"gopkg.in/urfave/cli.v1"
)

type EnvironmentCommand struct {
	environmentName string
	environmentAPI  api.EnvironmentAPI
	log             log.Logger
}

func (cmd *EnvironmentCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
}

func (cmd *EnvironmentCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environmentName == "" {
		return fmt.Errorf("Environment name is required")
	}

	if !ctx.Project.HasEnvironment(cmd.environmentName) {
		return fmt.Errorf("No environment named '%s' was found", cmd.environmentName)
	}

	return nil
}

func (cmd *EnvironmentCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *EnvironmentCommand) Environment(ctx *ecso.CommandContext) *ecso.Environment {
	return ctx.Project.Environments[cmd.environmentName]
}
