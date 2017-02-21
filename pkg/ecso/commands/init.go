package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

func NewInitCommand(projectName string) ecso.Command {
	return &initCommand{
		projectName: projectName,
	}
}

type initCommand struct {
	projectName string
}

func (cmd *initCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
}

func (cmd *initCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log = ctx.Config.Logger()
	)

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	project := ecso.NewProject(wd, cmd.projectName)

	if err := os.MkdirAll(filepath.Join(project.Dir(), ".ecso"), os.ModePerm); err != nil {
		return err
	}

	if err := project.Save(); err != nil {
		return err
	}

	log.Infof("Created project file at %s", project.ProjectFile())
	ui.BannerGreen(log, "Successfully created project '%s'.", project.Name)

	return nil
}

func (cmd *initCommand) Prompt(ctx *ecso.CommandContext) error {
	log := ctx.Config.Logger()

	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	ui.BannerBlue(log, "Creating a new ecso project")

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
