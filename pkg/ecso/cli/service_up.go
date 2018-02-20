package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		Environment cli.StringFlag
	}{
		Environment: cli.StringFlag{
			Name:   "environment",
			Usage:  "The name of the environment to deploy to",
			EnvVar: "ECSO_ENVIRONMENT",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceUpCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
		})
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploy a service",
		Description: "The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.Environment,
		},
	}
}
