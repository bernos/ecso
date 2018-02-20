package cli

import (
	"fmt"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceLsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "Environment to query",
			EnvVar: "ECSO_ENVIRONMENT",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		e := ctx.String(flags.Environment.Name)

		if e == "" {
			e = os.Getenv("ECSO_ENVIRONMENT")
		}

		if e == "" {
			return nil, ecso.NewArgumentRequiredError("environment")
		}

		if !project.HasEnvironment(e) {
			return nil, fmt.Errorf("Environment '%s' does not exist in the project", e)
		}

		return commands.NewServiceLsCommand(e, cfg.EnvironmentAPI(project.Environments[e].Region)), nil
	}

	return cli.Command{
		Name:      "ls",
		Usage:     "List services",
		ArgsUsage: "",
		Action:    MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
		},
	}
}
