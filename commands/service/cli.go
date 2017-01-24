package service

import (
	"github.com/bernos/ecso/commands/service/addservice"
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
		},
	}
}
