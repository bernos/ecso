package env

import (
	"fmt"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Unset string
}{
	Unset: "unset",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "env",
		Usage:     "Output shell environment configuration for an ecso environment",
		ArgsUsage: "[environment]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Unset,
				Usage: "If set, output shell commands to unset all ecso environment variables",
			},
		},
		Action: commands.MakeAction(FromCliContext, dispatcher),
	}
}

func FromCliContext(c *cli.Context) ecso.Command {
	return New(c.Args().First(), func(opt *Options) {
		opt.Unset = c.Bool(keys.Unset)
	})
}

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

func (cmd *envCommand) Execute(project *ecso.Project, cfg *ecso.Config, prefs ecso.UserPreferences) error {
	if cmd.options.EnvironmentName != "" {
		if _, ok := project.Environments[cmd.options.EnvironmentName]; ok {
			if cmd.options.Unset {
				fmt.Printf("unset ECSO_ENVIRONMENT\n")
			} else {
				fmt.Printf("export ECSO_ENVIRONMENT=%s\n", cmd.options.EnvironmentName)
			}
		}
	}
	return nil
}
