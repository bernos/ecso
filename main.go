package main

import (
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/cli"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/log"
)

var (
	logger  = log.NewLogger(os.Stderr, "")
	version = "0.0.0"
)

func main() {
	project := MustLoadProject(ecso.LoadCurrentProject())
	cfg := MustLoadConfig(config.NewConfig(version))
	prefs := MustLoadUserPreferences(ecso.LoadCurrentUserPreferences())
	dispatcher := ecso.NewDispatcher(project, cfg, prefs)
	app := cli.NewApp(cfg, dispatcher)

	err := app.Run(os.Args)

	if err != nil {
		ExitWithError(err, 1)
	}
}

func ExitWithError(err error, code int) {
	logger.Errorf(err.Error())
	os.Exit(code)
}

func MustLoadConfig(cfg *config.Config, err error) *config.Config {
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
