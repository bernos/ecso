package env

import (
	"fmt"

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
	if cmd.options.EnvironmentName != "" {
		if _, ok := ctx.Project.Environments[cmd.options.EnvironmentName]; ok {
			if cmd.options.Unset {
				fmt.Printf("unset ECSO_ENVIRONMENT\n")
			} else {
				fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.options.EnvironmentName)
			}
		}
	}
	return nil
}
