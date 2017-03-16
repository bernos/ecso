package main

import (
	"os"

	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var (
	log     = ecso.NewLogger(os.Stderr, "")
	version = "0.0.0"
)

func main() {
	project := MustLoadProject(ecso.LoadCurrentProject())
	cfg := MustLoadConfig(ecso.NewConfig())
	prefs := MustLoadUserPreferences(ecso.LoadCurrentUserPreferences())
	dispatcher := ecso.NewDispatcher(project, cfg, prefs, version)
	app := cmd.NewApp(version, dispatcher)

	cli.ErrWriter = cfg.Logger().ErrWriter()

	err := app.Run(os.Args)

	if err != nil {
		ExitWithError(err, 1)
	}
}

func ExitWithError(err error, code int) {
	log.Errorf(err.Error())
	os.Exit(code)
}

func MustLoadConfig(cfg *ecso.Config, err error) *ecso.Config {
	if err != nil {
		ExitWithError(err, 1)
	}
	return cfg
}

func MustLoadUserPreferences(prefs *ecso.UserPreferences, err error) *ecso.UserPreferences {
	if err != nil {
		ExitWithError(err, 1)
	}
	return prefs
}

func MustLoadProject(project *ecso.Project, err error) *ecso.Project {
	if err != nil {
		ExitWithError(err, 1)
	}
	return project
}
