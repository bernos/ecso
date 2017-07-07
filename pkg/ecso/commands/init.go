package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

func NewInitCommand(projectName string, log log.Logger) ecso.Command {
	return &initCommand{
		projectName: projectName,
		log:         log,
	}
}

type initCommand struct {
	projectName string
	log         log.Logger
}

func (cmd *initCommand) Execute(ctx *ecso.CommandContext) error {
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

	cmd.log.Infof("Created project file at %s", project.ProjectFile())
	ui.BannerGreen(cmd.log, "Successfully created project '%s'.", project.Name)

	return nil
}

func (cmd *initCommand) Prompt(ctx *ecso.CommandContext) error {
	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	ui.BannerBlue(cmd.log, "Creating a new ecso project")

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
