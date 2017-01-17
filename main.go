package main

import (
	"os"

	"github.com/bernos/ecso/commands/initcommand"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	log := ecso.NewLogger(os.Stdout)

	cli.ErrWriter = log.ErrWriter()

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
		{
			Name:      "init",
			Usage:     "Initialise a new ecso project",
			ArgsUsage: "project",
			Action: func(c *cli.Context) error {
				if err := initcommand.FromCliContext(c).Execute(log); err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
