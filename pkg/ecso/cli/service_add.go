package cli

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewServiceAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	flags := struct {
		DesiredCount cli.IntFlag
		Route        cli.StringFlag
		Port         cli.IntFlag
	}{

		DesiredCount: cli.IntFlag{
			Name:  "desired-cout",
			Usage: "The desired number of service instances",
		},
		Route: cli.StringFlag{
			Name:  "route",
			Usage: "If set, the service will be registered with the load balancer at this route",
		},
		Port: cli.IntFlag{
			Name:  "port",
			Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
		},
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewServiceAddCommand(ctx.Args().First()).
			WithDesiredCount(ctx.Int(flags.DesiredCount.Name)).
			WithRoute(ctx.String(flags.Route.Name)).
			WithPort(ctx.Int(flags.Port.Name)), nil
	}

	return cli.Command{
		Name:        "add",
		Usage:       "Adds a new service to the project",
		Description: "The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.",
		ArgsUsage:   "SERVICE",
		Action:      MakeAction(dispatcher, fn),
		Flags: []cli.Flag{
			flags.DesiredCount,
			flags.Route,
			flags.Port,
		},
	}
}
