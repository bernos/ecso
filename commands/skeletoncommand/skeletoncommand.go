package skeletoncommand

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Unset string
}{
	Unset: "unset",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:      "TODO",
		Usage:     "TODO",
		ArgsUsage: "[TODO]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Unset,
				Usage: "TODO",
			},
		},
		Action: commands.MakeAction(FromCliContext, dispatcher),
	}
}

func FromCliContext(c *cli.Context) ecso.Command {
	return New(c.Args().First(), func(opt *Options) {
		// TODO: populate options from c
	})
}

type Options struct {
	EnvironmentName string
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		EnvironmentName: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &skeletonCommand{
		options: o,
	}
}

type skeletonCommand struct {
	options *Options
}

func (cmd *skeletonCommand) Execute(project *ecso.Project, cfg *ecso.Config, prefs ecso.UserPreferences) error {
	return nil
}
