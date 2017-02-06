package ls

import (
	"github.com/bernos/ecso/cmd"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"

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
				Name:   keys.Environment,
				Usage:  "Environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: cmd.MakeAction(dispatcher, FromCliContext),
	}
}

func FromCliContext(c *cli.Context) (ecso.Command, error) {
	env := c.String(keys.Environment)

	if env == "" {
		return nil, cmd.NewOptionRequiredError(keys.Environment)
	}

	return commands.NewServiceLsCommand(env, func(opt *commands.ServiceLsOptions) {
		// TODO: populate options from c
	}), nil
}
