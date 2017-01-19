package initcommand

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

func New(projectName string, options ...func(*Options)) commands.Command {
	o := &Options{
		ProjectName: projectName,
	}

	for _, option := range options {
		option(o)
	}

	return &initCommand{
		options: o,
	}
}

type initCommand struct {
	options *Options
}

func (cmd *initCommand) Execute(cfg *ecso.Config) error {
	log := cfg.Logger

	projectFile, err := util.GetCurrentProjectFile()

	if err != nil {
		return err
	}

	if err := errIfProjectExists(projectFile); err != nil {
		return err
	}

	log.BannerBlue("Creating a new ecso project")

	if err := promptForMissingOptions(cmd.options); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(projectFile), os.ModePerm); err != nil {
		return err
	}

	if err := util.SaveCurrentProject(ecso.NewProject(cmd.options.ProjectName)); err != nil {
		return err
	}

	log.Infof("Created project file at %s", projectFile)
	log.BannerGreen("Successfully created project '%s'.", cmd.options.ProjectName)

	return nil
}

func errIfProjectExists(projectFile string) error {
	exists, err := dirExists(path.Dir(projectFile))

	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("Found an existing project at %s.", path.Dir(projectFile))
	}

	return nil
}

func promptForMissingOptions(options *Options) error {
	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	name := filepath.Base(wd)

	if err := ui.AskStringIfEmptyVar(
		&options.ProjectName,
		"What is the name of your project?",
		name,
		ui.ValidateNotEmpty("Project name is required")); err != nil {
		return err
	}

	return nil
}

func dirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}
