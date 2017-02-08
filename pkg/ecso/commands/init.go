package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type InitOptions struct {
	ProjectName string
}

func NewInitCommand(projectName string, options ...func(*InitOptions)) ecso.Command {
	o := &InitOptions{
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
	options *InitOptions
}

func (cmd *initCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log = ctx.Config.Logger()
	)

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
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

func (cmd *initCommand) Prompt(ctx *ecso.CommandContext) error {
	log := ctx.Config.Logger()

	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	log.BannerBlue("Creating a new ecso project")

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	return ui.AskStringIfEmptyVar(
		&cmd.options.ProjectName,
		"What is the name of your project?",
		filepath.Base(wd),
		ui.ValidateNotEmpty("Project name is required"))
}

func (cmd *initCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
