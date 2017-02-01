package env

import (
	"fmt"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
)

// Options for the env command
type Options struct {
	EnvironmentName string
	Unset           bool
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &envCommand{
		options: o,
	}
}

type envCommand struct {
	options *Options
}

func (cmd *envCommand) Execute(ctx *ecso.CommandContext) error {
	if cmd.options.Unset {
		oldPS1 := os.Getenv("ECSO_OLD_PS1")

		fmt.Printf("unset ECSO_ENVIRONMENT; ")
		fmt.Printf("unset ECSO_OLD_PS1; ")

		if oldPS1 != "" {
			fmt.Printf("export PS1=\"%s\"\n", oldPS1)
		}
	} else if cmd.options.EnvironmentName != "" {
		if _, ok := ctx.Project.Environments[cmd.options.EnvironmentName]; ok {

			ps1 := os.Getenv("PS1")

			if ps1 != "" {
				fmt.Printf("export PS1=\"%s[ecso::%s:%s]> \"\n", ps1, ctx.Project.Name, cmd.options.EnvironmentName)
				fmt.Printf("export ECSO_OLD_PS1=\"%s\"\n", ps1)
			}

			fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.options.EnvironmentName)
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
