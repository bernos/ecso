package commands

import (
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

func NewEnvironmentDownCommand(environmentName string, environmentAPI api.EnvironmentAPI) *EnvironmentDownCommand {
	return &EnvironmentDownCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

type EnvironmentDownCommand struct {
	*EnvironmentCommand
	force bool
}

func (cmd *EnvironmentDownCommand) WithForce(force bool) *EnvironmentDownCommand {
	cmd.force = force
	return cmd
}

func (cmd *EnvironmentDownCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		project = ctx.Project
		env     = cmd.Environment(ctx)
	)

	return cmd.environmentAPI.EnvironmentDown(project, env, w)
}

func (cmd *EnvironmentDownCommand) Validate(ctx *ecso.CommandContext) error {
	if err := cmd.EnvironmentCommand.Validate(ctx); err != nil {
		return err
	}

	if !cmd.force {
		return ecso.NewOptionRequiredError("force")
	}

	return nil
}
