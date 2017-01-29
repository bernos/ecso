package initcommand

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	return New(c.Args().First()), nil
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "init",
		Usage:     "Initialise a new ecso project",
		ArgsUsage: "[project]",
		Action:    commands.MakeAction(dispatcher, FromCliContext, ecso.SkipEnsureProjectExists()),
	}
}
