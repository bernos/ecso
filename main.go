package main

import (
	"os"

	"github.com/bernos/ecso/commands/addenvironmentcommand"
	"github.com/bernos/ecso/commands/envcommand"
	"github.com/bernos/ecso/commands/environmentupcommand"
	"github.com/bernos/ecso/commands/initcommand"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var log = ecso.NewLogger(os.Stdout)

func main() {
	project := MustLoadProject(ecso.LoadCurrentProject())
	cfg := MustLoadConfig(ecso.NewConfig())
	prefs := MustLoadUserPreferences(ecso.LoadUserPreferences())
	dispatcher := ecso.NewDispatcher(project, cfg, prefs)

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
		initcommand.CliCommand(dispatcher),
		addenvironmentcommand.CliCommand(dispatcher),
		envcommand.CliCommand(dispatcher),
		environmentupcommand.CliCommand(dispatcher),
	}

	app.Run(os.Args)
}

func MustLoadConfig(cfg *ecso.Config, err error) *ecso.Config {
	if err != nil {
		log.Fatalf(err.Error())
	}
	return cfg
}

func MustLoadUserPreferences(prefs ecso.UserPreferences, err error) ecso.UserPreferences {
	if err != nil {
		log.Fatalf(err.Error())
	}
	return prefs
}

func MustLoadProject(project *ecso.Project, err error) *ecso.Project {
	if err != nil {
		log.Fatalf(err.Error())
	}
	return project
}
