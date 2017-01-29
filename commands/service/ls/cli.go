package ls

import (
	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Environment string
}{
	Environment: "environment",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "ls",
		Usage: "List services",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Environment,
				Usage: "Environment to query",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.String(keys.Environment)

	if env == "" {
		return nil, commands.NewOptionRequiredError(keys.Environment)
	}

	return New(env, func(opt *Options) {
		// TODO: populate options from c
	}), nil
}
