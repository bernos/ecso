package main

import (
	"os"

	"github.com/bernos/ecso/commands/addenvironmentcommand"
	"github.com/bernos/ecso/commands/initcommand"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	cfg := ecso.NewConfig()

	cli.ErrWriter = cfg.Logger.ErrWriter()

	app := cli.NewApp()
	app.Name = "ecso"
	app.Usage = "Manage Amazon ECS projects"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Brendan McMahon",
			Email: "bernos@gmail.com",
		},
	}

	app.Commands = []cli.Command{
		initcommand.CliCommand(cfg),
		addenvironmentcommand.CliCommand(cfg),
	}

	app.Run(os.Args)
}
