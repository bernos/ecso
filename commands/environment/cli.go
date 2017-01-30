package environment

import (
	"github.com/bernos/ecso/commands/environment/addenvironment"
	"github.com/bernos/ecso/commands/environment/environmentup"
	"github.com/bernos/ecso/commands/environment/rm"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "environment",
		Usage: "Manage ecso environments",
		Subcommands: []cli.Command{
			addenvironment.CliCommand(dispatcher),
			environmentup.CliCommand(dispatcher),
			rm.CliCommand(dispatcher),
		},
	}
}
