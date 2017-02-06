package initcommand

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	return commands.NewInitCommand(c.Args().First()), nil
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "init",
		Usage:     "Initialise a new ecso project",
		ArgsUsage: "[project]",
		Action:    cmd.MakeAction(dispatcher, FromCliContext, ecso.SkipEnsureProjectExists()),
	}
}
