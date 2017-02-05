package rm

import (
	"os"

	"github.com/bernos/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Force string
}{
	Force: "force",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:        "rm",
		Usage:       "Removes an ecso environment",
		Description: "TODO",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Force,
				Usage: "Required. Confirms the environment will be removed",
			},
		},
		Action: commands.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	force := c.Bool(keys.Force)
	env := c.Args().First()

	if env == "" {
		env = os.Getenv("ECSO_ENVIRONMENT")
	}

	if env == "" {
		return nil, commands.NewArgumentRequiredError("environment")
	}

	if !force {
		return nil, commands.NewOptionRequiredError(keys.Force)
	}

	return New(env, func(opt *Options) {
		// TODO: populate options from c
	}), nil
}
