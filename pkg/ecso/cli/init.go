package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
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
