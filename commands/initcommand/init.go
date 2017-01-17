package initcommand

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
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

func (cmd *initCommand) Execute(log ecso.Logger) error {

	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	projectFile := filepath.Join(wd, ".ecso", "project.json")

	exists, err := dirExists(path.Dir(projectFile))

	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("Found an existing project at %s.", path.Dir(projectFile))
	}

	log.BannerBlue("Creating a new ecso project")

	if err := os.MkdirAll(path.Dir(projectFile), os.ModePerm); err != nil {
		return err
	}

	project := ecso.NewProject(cmd.options.ProjectName)

	file, err := os.Create(projectFile)

	if err != nil {
		return err
	}

	if err := project.Save(file); err != nil {
		return err
	}

	log.Infof("Created project file at %s\n", projectFile)
	log.BannerGreen("Successfully created project '%s'.", project.Name)

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
