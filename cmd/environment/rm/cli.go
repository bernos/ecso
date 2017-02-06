package rm

import (
	"os"

	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"

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
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	force := c.Bool(keys.Force)
	env := c.Args().First()

	if env == "" {
		env = os.Getenv("ECSO_ENVIRONMENT")
	}

	if env == "" {
		return nil, cmd.NewArgumentRequiredError("environment")
	}

	if !force {
		return nil, cmd.NewOptionRequiredError(keys.Force)
	}

	return commands.NewEnvironmentRmCommand(env, func(opt *commands.EnvironmentRmOptions) {
		// TODO: populate options from c
	}), nil
}
