package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
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

func (cmd *initCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	green := ui.NewBannerWriter(w, ui.GreenBold)
	info := ui.NewInfoWriter(w)

	if err := cmd.prompt(ctx, r, w); err != nil {
		return err
	}

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	project := ecso.NewProject(wd, cmd.projectName, ctx.EcsoVersion)

	if err := os.MkdirAll(project.DotDir(), os.ModePerm); err != nil {
		return err
	}

	if err := project.Save(); err != nil {
		return err
	}

	fmt.Fprintf(info, "Created project file at %s", project.ProjectFile())
	fmt.Fprintf(green, "Successfully created project '%s'.", project.Name)

	return nil
}

func (cmd *initCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *initCommand) prompt(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	blue := ui.NewBannerWriter(w, ui.BlueBold)

	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	fmt.Fprint(blue, "Creating a new ecso project")

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	return ui.AskStringIfEmptyVar(
		r, w,
		&cmd.projectName,
		"What is the name of your project?",
		filepath.Base(wd),
		ui.ValidateNotEmpty("Project name is required"))
}
