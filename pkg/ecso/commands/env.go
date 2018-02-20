package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
)

func NewEnvCommand(environmentName string) *EnvCommand {
	return &EnvCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
		},
	}
}

type EnvCommand struct {
	*EnvironmentCommand
	unset bool
}

func (cmd *EnvCommand) WithUnset(unset bool) *EnvCommand {
	cmd.unset = unset
	return cmd
}

func (cmd *EnvCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	if cmd.unset {
		oldPS1 := os.Getenv("ECSO_OLD_PS1")

		fmt.Printf("unset ECSO_ENVIRONMENT; ")
		fmt.Printf("unset ECSO_OLD_PS1; ")

		if oldPS1 != "" {
			fmt.Printf("export PS1=\"%s\"\n", oldPS1)
		}
	} else if cmd.environmentName != "" {
		if _, ok := ctx.Project.Environments[cmd.environmentName]; ok {

			ps1 := os.Getenv("PS1")

			if ps1 != "" {
				fmt.Printf("export PS1=\"%s $(tput setaf 2)[ecso::%s:%s]:$(tput sgr0) \"\n", ps1, ctx.Project.Name, cmd.environmentName)
				fmt.Printf("export ECSO_OLD_PS1=\"%s\"\n", ps1)
			}

			fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.environmentName)
		}
	}
	return nil
}

func (cmd *EnvCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
