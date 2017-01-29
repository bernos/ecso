package env

import (
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
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.Args().First()

	if env == "" {
		return nil, commands.NewArgumentRequiredError("environment")
	}

	return New(env, func(opt *Options) {
		opt.Unset = c.Bool(keys.Unset)
	}), nil
}
