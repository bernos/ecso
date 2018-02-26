package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
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
	wd, err := ecso.GetCurrentProjectDir()
	if err != nil {
		return err
	}

	project := ecso.NewProject(wd, cmd.projectName, ctx.EcsoVersion)

	if err := os.MkdirAll(project.DotDir(), os.ModePerm); err != nil {
		return err
	}

	return project.Save()
}

func (cmd *initCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.projectName == "" {
		return fmt.Errorf("Project name required")
	}
	return nil
}
