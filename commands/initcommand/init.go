package initcommand

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func New(projectName string, options ...func(*Options)) ecso.Command {
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

func (cmd *initCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log = ctx.Config.Logger
	)

	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	log.BannerBlue("Creating a new ecso project")

	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err := promptForMissingOptions(cmd.options); err != nil {
		return err
	}

	project := ecso.NewProject(wd, cmd.options.ProjectName)

	if err := os.MkdirAll(filepath.Join(project.Dir(), ".ecso"), os.ModePerm); err != nil {
		return err
	}

	if err := project.Save(); err != nil {
		return err
	}

	log.Infof("Created project file at %s", project.ProjectFile())
	log.BannerGreen("Successfully created project '%s'.", project.Name)

	return nil
}

func promptForMissingOptions(options *Options) error {
	wd, err := ecso.GetCurrentProjectDir()

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
