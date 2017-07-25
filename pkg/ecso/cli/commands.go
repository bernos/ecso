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

func NewEnvCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewEnvCommand(ctx.Args().First()).
			WithUnset(ctx.Bool(commands.EnvUnsetOption)), nil
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

func NewEnvironmentCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "environment",
		Usage: "Manage ecso environments",
		Subcommands: []cli.Command{
			NewEnvironmentAddCliCommand(project, dispatcher),
			NewEnvironmentPsCliCommand(project, dispatcher),
			NewEnvironmentUpCliCommand(project, dispatcher),
			NewEnvironmentRmCliCommand(project, dispatcher),
			NewEnvironmentDescribeCliCommand(project, dispatcher),
			NewEnvironmentDownCliCommand(project, dispatcher),
		},
	}
}

func NewEnvironmentAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		region := ctx.String(commands.EnvironmentAddRegionOption)
		if region == "" {
			region = "ap-southeast-2"
		}

		return commands.NewEnvironmentAddCommand(ctx.Args().First(), cfg.EnvironmentAPI(region)).
			WithALBSubnets(ctx.String(commands.EnvironmentAddALBSubnetsOption)).
			WithInstanceSubnets(ctx.String(commands.EnvironmentAddInstanceSubnetsOption)).
			WithInstanceType(ctx.String(commands.EnvironmentAddInstanceTypeOption)).
			WithRegion(ctx.String(commands.EnvironmentAddRegionOption)).
			WithSize(ctx.Int(commands.EnvironmentAddSizeOption)).
			WithVPCID(ctx.String(commands.EnvironmentAddVPCOption)), nil
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

func NewEnvironmentDescribeCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentDescribeCommand(env.Name, cfg.EnvironmentAPI(env.Region))
		})
	}

	return cli.Command{
		Name:      "describe",
		Usage:     "Describes an ecso environment",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentDownCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentDownCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithForce(ctx.Bool(commands.EnvironmentDownForceOption))
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

func NewEnvironmentPsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentPsCommand(env.Name, cfg.EnvironmentAPI(env.Region))
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

func NewEnvironmentRmCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentRmCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithForce(ctx.Bool(commands.EnvironmentRmForceOption))
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

func NewEnvironmentUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentUpCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithDryRun(ctx.Bool(commands.EnvironmentUpDryRunOption)).
				WithForce(ctx.Bool(commands.EnvironmentUpForceOption))
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

func NewInitCliCommand(project *ecso.Project, d dispatcher.Dispatcher) cli.Command {
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

func NewServiceCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Manage ecso services",
		Subcommands: []cli.Command{
			NewServiceAddCliCommand(project, dispatcher),
			NewServiceUpCliCommand(project, dispatcher),
			NewServiceDownCliCommand(project, dispatcher),
			NewServiceLsCliCommand(project, dispatcher),
			NewServicePsCliCommand(project, dispatcher),
			NewServiceEventsCliCommand(project, dispatcher),
			NewServiceLogsCliCommand(project, dispatcher),
			NewServiceDescribeCliCommand(project, dispatcher),
			NewServiceRollbackCliCommand(project, dispatcher),
			NewServiceVersionsCliCommand(project, dispatcher),
		},
	}
}

func NewServiceAddCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewServiceAddCommand(ctx.Args().First()).
			WithDesiredCount(ctx.Int(commands.ServiceAddDesiredCountOption)).
			WithRoute(ctx.String(commands.ServiceAddRouteOption)).
			WithPort(ctx.Int(commands.ServiceAddPortOption)), nil
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

func NewServiceDescribeCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceDescribeCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
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

func NewServiceDownCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceDownCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region)).
				WithForce(ctx.Bool(commands.ServiceDownForceOption))
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

func NewServiceEventsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceEventsCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
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

func NewServiceLogsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceLogsCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
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

func NewServiceLsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		e := ctx.String(commands.ServiceLsEnvironmentOption)

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

func NewServicePsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServicePsCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
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

func NewServiceRollbackCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceRollbackCommand(
				service.Name,
				env.Name,
				ctx.String(commands.ServiceRollbackVersionOption),
				cfg.ServiceAPI(env.Region))
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

func NewServiceUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
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

func NewServiceVersionsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceVersionsCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region))
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

func makeEnvironmentCommand(c *cli.Context, project *ecso.Project, fn func(*ecso.Environment) ecso.Command) (ecso.Command, error) {
	name := c.Args().First()

	if name == "" {
		name = os.Getenv("ECSO_ENVIRONMENT")
	}

	if name == "" {
		return nil, ecso.NewArgumentRequiredError("environment")
	}

	if !project.HasEnvironment(name) {
		return nil, fmt.Errorf("Environment '%s' does not exist in the project", name)
	}

	return fn(project.Environments[name]), nil
}

func makeServiceCommand(c *cli.Context, project *ecso.Project, fn func(*ecso.Service, *ecso.Environment) ecso.Command) (ecso.Command, error) {
	name := c.Args().First()
	environmentName := c.String(commands.ServiceEnvironmentOption)

	if name == "" {
		return nil, ecso.NewArgumentRequiredError("service")
	}

	if environmentName == "" {
		return nil, ecso.NewOptionRequiredError("environment")
	}

	if !project.HasService(name) {
		return nil, fmt.Errorf("Service '%s' does not exist in the project", name)
	}

	if !project.HasEnvironment(environmentName) {
		return nil, fmt.Errorf("Environment '%s' does not exist in the project", name)
	}

	return fn(project.Services[name], project.Environments[environmentName]), nil
}
