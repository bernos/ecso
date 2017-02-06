package rm

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"

	"gopkg.in/urfave/cli.v1"
)

var keys = struct {
	Name  string
	Force string
}{
	Name:  "name",
	Force: "force",
}

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "rm",
		Usage: "Removes an entire environment",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Name,
				Usage:  "The name of the environment to remove",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.BoolFlag{
				Name:  keys.Force,
				Usage: "Required. Confirms the environment will be removed",
			},
		},
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.String(keys.Name)
	force := c.Bool(keys.Force)

	if env == "" {
		return nil, cmd.NewOptionRequiredError(keys.Name)
	}

	if !force {
		return nil, cmd.NewOptionRequiredError(keys.Force)
	}

	return commands.NewEnvironmentRmCommand(env, func(opt *commands.EnvironmentRmOptions) {
		// TODO: populate options from c
	}), nil
}
