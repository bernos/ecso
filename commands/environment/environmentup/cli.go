package environmentup

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name   string
	DryRun string
}{
	Name:   "name",
	DryRun: "dry-run",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "up",
		Usage: "Bring up the named environment",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Name,
				Usage:  "The name of the environment to bring up. If the environment doesn't exist it will be created, otherwise it will be updated.",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.BoolFlag{
				Name:  keys.DryRun,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
		},
		Action: commands.MakeAction(FromCliContext, dispatcher),
	}
}

func FromCliContext(c *cli.Context) ecso.Command {
	return New(c.String(keys.Name), func(opt *Options) {
		opt.DryRun = c.Bool(keys.DryRun)
	})
}
