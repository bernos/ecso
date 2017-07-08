package cli

import (
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"github.com/bernos/ecso/pkg/ecso/config"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

func NewEnvCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewEnvCommand(ctx.Args().First()), nil
	}

	return cli.Command{
		Name:      "env",
		Usage:     "Display the commands to set up the default environment for the ecso cli tool",
		ArgsUsage: "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  commands.EnvUnsetOption,
				Usage: "If set, output shell commands to unset all ecso environment variables",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "environment",
		Usage: "Manage ecso environments",
		Subcommands: []cli.Command{
			NewEnvironmentAddCliCommand(dispatcher),
			NewEnvironmentPsCliCommand(dispatcher),
			NewEnvironmentUpCliCommand(dispatcher),
			NewEnvironmentRmCliCommand(dispatcher),
			NewEnvironmentDescribeCliCommand(dispatcher),
			NewEnvironmentDownCliCommand(dispatcher),
		},
	}
}

func NewEnvironmentAddCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewEnvironmentAddCommand(ctx.Args().First(), cfg.EnvironmentAPI()), nil
	}

	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[ENVIRONMENT]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  commands.EnvironmentAddVPCOption,
				Usage: "The vpc to create the environment in",
			},
			cli.StringFlag{
				Name:  commands.EnvironmentAddALBSubnetsOption,
				Usage: "The subnets to place the application load balancer in",
			},
			cli.StringFlag{
				Name:  commands.EnvironmentAddInstanceSubnetsOption,
				Usage: "The subnets to place the ecs container instances in",
			},
			cli.StringFlag{
				Name:  commands.EnvironmentAddRegionOption,
				Usage: "The AWS region to create the environment in",
			},
			cli.IntFlag{
				Name:  commands.EnvironmentAddSizeOption,
				Usage: "Then number of container instances to create",
			},
			cli.StringFlag{
				Name:  commands.EnvironmentAddInstanceTypeOption,
				Usage: "The type of container instances to create",
			},
			cli.StringFlag{
				Name:  commands.EnvironmentAddKeyPairOption,
				Usage: "The keypair to use when accessing EC2 instances",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentDescribeCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, func(name string) ecso.Command {
			return commands.NewEnvironmentDescribeCommand(name, cfg.EnvironmentAPI())
		})
	}

	return cli.Command{
		Name:      "describe",
		Usage:     "Describes an ecso environment",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentDownCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, func(name string) ecso.Command {
			return commands.NewEnvironmentDownCommand(name, cfg.EnvironmentAPI())
		})
	}

	return cli.Command{
		Name:        "down",
		Usage:       "Terminates an ecso environment",
		Description: "Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  commands.EnvironmentDownForceOption,
				Usage: "Required. Confirms the environment will be terminated",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentPsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, func(name string) ecso.Command {
			return commands.NewEnvironmentPsCommand(name, cfg.EnvironmentAPI())
		})
	}

	return cli.Command{
		Name:        "ps",
		Usage:       "Lists containers running in an environment",
		Description: "",
		ArgsUsage:   "ENVIRONMENT",
		Action:      MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentRmCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, func(name string) ecso.Command {
			return commands.NewEnvironmentRmCommand(name, cfg.EnvironmentAPI())
		})
	}

	return cli.Command{
		Name:        "rm",
		Usage:       "Removes an ecso environment",
		Description: "Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  commands.EnvironmentRmForceOption,
				Usage: "Required. Confirms the environment will be removed",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentUpCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, func(name string) ecso.Command {
			return commands.NewEnvironmentUpCommand(name, cfg.EnvironmentAPI())
		})
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploys the infrastructure for an ecso environment",
		Description: "All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  commands.EnvironmentUpDryRunOption,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
			cli.BoolFlag{
				Name:  commands.EnvironmentUpForceOption,
				Usage: "Override warnings about first time environment deployments if cloud formation stack already exists",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewInitCliCommand(d dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewInitCommand(ctx.Args().First()), nil
	}

	return cli.Command{
		Name:        "init",
		Usage:       "Initialise a new ecso project",
		Description: "Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.",
		ArgsUsage:   "[PROJECT]",
		Action:      MakeAction(d, fn, dispatcher.SkipEnsureProjectExists()),
	}
}

func NewServiceCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Manage ecso services",
		Subcommands: []cli.Command{
			NewServiceAddCliCommand(dispatcher),
			NewServiceUpCliCommand(dispatcher),
			NewServiceDownCliCommand(dispatcher),
			NewServiceLsCliCommand(dispatcher),
			NewServicePsCliCommand(dispatcher),
			NewServiceEventsCliCommand(dispatcher),
			NewServiceLogsCliCommand(dispatcher),
			NewServiceDescribeCliCommand(dispatcher),
			NewServiceRollbackCliCommand(dispatcher),
			NewServiceVersionsCliCommand(dispatcher),
		},
	}
}

func NewServiceAddCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewServiceAddCommand(ctx.Args().First()), nil
	}

	return cli.Command{
		Name:        "add",
		Usage:       "Adds a new service to the project",
		Description: "The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  commands.ServiceAddDesiredCountOption,
				Usage: "The desired number of service instances",
			},
			cli.StringFlag{
				Name:  commands.ServiceAddRouteOption,
				Usage: "If set, the service will be registered with the load balancer at this route",
			},
			cli.IntFlag{
				Name:  commands.ServiceAddPortOption,
				Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceDescribeCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceDescribeCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:        "describe",
		Usage:       "Lists details of a deployed service",
		Description: "Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceDownCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceDownCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:        "down",
		Usage:       "terminates a service",
		Description: "The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.BoolFlag{
				Name:  commands.ServiceDownForceOption,
				Usage: "Required. Confirms the service will be terminated",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceEventsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceEventsCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:      "events",
		Usage:     "List ECS events for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLogsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceLogsCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:      "logs",
		Usage:     "output service logs",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		e := ctx.String(commands.ServiceLsEnvironmentOption)

		if e == "" {
			e = os.Getenv("ECSO_ENVIRONMENT")
		}

		if e == "" {
			return nil, ecso.NewArgumentRequiredError("environment")
		}

		return commands.NewServiceLsCommand(e, cfg.EnvironmentAPI()), nil
	}

	return cli.Command{
		Name:      "ls",
		Usage:     "List services",
		ArgsUsage: "",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceLsEnvironmentOption,
				Usage:  "Environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServicePsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServicePsCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:      "ps",
		Usage:     "Show running tasks for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceRollbackCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceRollbackCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:        "rollback",
		Usage:       "Rollback a service to an earlier version",
		Description: "Replace the currently running service with a previously deployed service version",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.StringFlag{
				Name:  commands.ServiceRollbackVersionOption,
				Usage: "The version to rollback to",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceUpCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceUpCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploy a service",
		Description: "The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceVersionsCliCommand(dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, func(name string) ecso.Command {
			return commands.NewServiceVersionsCommand(name, cfg.ServiceAPI())
		})
	}

	return cli.Command{
		Name:      "versions",
		Usage:     "Show available versions for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEnvironmentOption,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func makeEnvironmentCommand(c *cli.Context, fn func(string) ecso.Command) (ecso.Command, error) {
	name := c.Args().First()

	if name == "" {
		name = os.Getenv("ECSO_ENVIRONMENT")
	}

	if name == "" {
		return nil, ecso.NewArgumentRequiredError("environment")
	}

	return fn(name), nil
}

func makeServiceCommand(c *cli.Context, fn func(string) ecso.Command) (ecso.Command, error) {
	name := c.Args().First()

	if name == "" {
		return nil, ecso.NewArgumentRequiredError("service")
	}

	return fn(name), nil
}
