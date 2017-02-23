package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

type EnvironmentCommand struct {
	environmentName string
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
