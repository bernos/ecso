package commands

import (
	"fmt"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
)

// Options for the env command
type EnvOptions struct {
	Environment string
	Unset       bool
}

func NewEnvCommand(environment string, options ...func(*EnvOptions)) ecso.Command {
	o := &EnvOptions{
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &envCommand{
		options: o,
	}
}

type envCommand struct {
	options *EnvOptions
}

func (cmd *envCommand) Execute(ctx *ecso.CommandContext) error {
	if cmd.options.Unset {
		oldPS1 := os.Getenv("ECSO_OLD_PS1")

		fmt.Printf("unset ECSO_ENVIRONMENT; ")
		fmt.Printf("unset ECSO_OLD_PS1; ")

		if oldPS1 != "" {
			fmt.Printf("export PS1=\"%s\"\n", oldPS1)
		}
	} else if cmd.options.Environment != "" {
		if _, ok := ctx.Project.Environments[cmd.options.Environment]; ok {

			ps1 := os.Getenv("PS1")

			if ps1 != "" {
				fmt.Printf("export PS1=\"%s $(tput setaf 2)[ecso::%s:%s]:$(tput sgr0) \"\n", ps1, ctx.Project.Name, cmd.options.Environment)
				fmt.Printf("export ECSO_OLD_PS1=\"%s\"\n", ps1)
			}

			fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.options.Environment)
		}
	}
	return nil
}

func (cmd *envCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *envCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
