package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

func NewInitCliCommand(project *ecso.Project, d dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewInitCommand(ctx.Args().First()), nil
	}

	return cli.Command{
		Name:        "init",
		Usage:       "Initialise a new ecso project",
		Description: "Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.",
		ArgsUsage:   "[PROJECT]",
		Action:      MakeAction(d, fn, dispatcher.SkipEnsureProjectExists()),
	}
}

type initCommandWrapper struct {
	cliCtx *cli.Context
}

func (wrapper *initCommandWrapper) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	var (
		projectName = wrapper.cliCtx.Args().First()
		blue        = ui.NewBannerWriter(w, ui.BlueBold)
		green       = ui.NewBannerWriter(w, ui.GreenBold)
		info        = ui.NewInfoWriter(w)
	)

	if ctx.Project != nil {
		return fmt.Errorf("Found an existing project at %s.", ctx.Project.ProjectFile())
	}

	fmt.Fprint(blue, "Creating a new ecso project")

	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	if err := ui.AskStringIfEmptyVar(
		r, w,
		&projectName,
		"What is the name of your project?",
		filepath.Base(wd),
		ui.ValidateNotEmpty("Project name is required")); err != nil {
		return err
	}

	cmd := commands.NewInitCommand(projectName)

	if err := cmd.Execute(ctx, r, w); err != nil {
		return err
	}

	fmt.Fprintf(info, "Created project file at %s", ctx.Project.ProjectFile())
	fmt.Fprintf(green, "Successfully created project '%s'.", ctx.Project.Name)

	return nil
}
