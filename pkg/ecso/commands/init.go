package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewInitCommand(projectName string) ecso.Command {
	return &initCommand{
		projectName: projectName,
	}
}

type initCommand struct {
	projectName string
}

func (cmd *initCommand) Execute(ctx *ecso.CommandContext, l log.Logger) error {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	project := ecso.NewProject(wd, cmd.projectName, ctx.EcsoVersion)

	if err := os.MkdirAll(filepath.Join(project.Dir(), ".ecso"), os.ModePerm); err != nil {
		return err
	}

	if err := project.Save(); err != nil {
		return err
	}

	l.Infof("Created project file at %s", project.ProjectFile())
	ui.BannerGreen(l, "Successfully created project '%s'.", project.Name)

	return nil
}

func (cmd *initCommand) Prompt(ctx *ecso.CommandContext, l log.Logger) error {
	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	ui.BannerBlue(l, "Creating a new ecso project")

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	return ui.AskStringIfEmptyVar(
		&cmd.projectName,
		"What is the name of your project?",
		filepath.Base(wd),
		ui.ValidateNotEmpty("Project name is required"))
}

func (cmd *initCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
