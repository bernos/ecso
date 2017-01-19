package initcommand

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

func FromCliContext(c *cli.Context) commands.Command {
	return New(c.Args().First())
}

func CliCommand(cfg *ecso.Config) cli.Command {
	return cli.Command{
		Name:      "init",
		Usage:     "Initialise a new ecso project",
		ArgsUsage: "[project]",
		Action: func(c *cli.Context) error {
			if err := FromCliContext(c).Execute(cfg); err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return nil
		},
	}
}
