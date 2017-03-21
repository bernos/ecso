package cmd

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

func NewApp(version string, dispatcher ecso.Dispatcher) *cli.App {
	app := cli.NewApp()
	app.Name = "ecso"
	app.Usage = "Manage Amazon ECS projects"
	app.Version = version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Brendan McMahon",
			Email: "bernos@gmail.com",
		},
	}

	cliDispatcher := CliDispatcher(dispatcher)

	app.Commands = []cli.Command{
		NewInitCliCommand(cliDispatcher),
		NewEnvironmentCliCommand(cliDispatcher),
		NewServiceCliCommand(cliDispatcher),
		NewEnvCliCommand(cliDispatcher),
	}

	return app
}
