package service

import (
	"github.com/bernos/ecso/commands/service/addservice"
	"github.com/bernos/ecso/commands/service/events"
	"github.com/bernos/ecso/commands/service/logs"
	"github.com/bernos/ecso/commands/service/ls"
	"github.com/bernos/ecso/commands/service/ps"
	"github.com/bernos/ecso/commands/service/purgedns"
	"github.com/bernos/ecso/commands/service/servicedown"
	"github.com/bernos/ecso/commands/service/serviceup"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Manage ecso services",
		Subcommands: []cli.Command{
			addservice.CliCommand(dispatcher),
			serviceup.CliCommand(dispatcher),
			servicedown.CliCommand(dispatcher),
			ls.CliCommand(dispatcher),
			ps.CliCommand(dispatcher),
			events.CliCommand(dispatcher),
			logs.CliCommand(dispatcher),
			purgedns.CliCommand(dispatcher),
		},
	}
}
