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
	options := struct {
		Unset string
	}{
		Unset: "unset",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewEnvCommand(ctx.Args().First()).
			WithUnset(ctx.Bool(options.Unset)), nil
	}

	return cli.Command{
		Name:      "env",
		Usage:     "Display the commands to set up the default environment for the ecso cli tool",
		ArgsUsage: "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  options.Unset,
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
	options := struct {
		VPC             string
		ALBSubnets      string
		InstanceSubnets string
		InstanceType    string
		Region          string
		Size            string
		KeyPair         string
	}{
		VPC:             "vpc",
		ALBSubnets:      "alb-subnets",
		InstanceSubnets: "instance-subnets",
		InstanceType:    "instance-type",
		Region:          "region",
		Size:            "size",
		KeyPair:         "keypair",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		region := ctx.String(options.Region)

		if region == "" {
			region = "ap-southeast-2"
		}

		return commands.NewEnvironmentAddCommand(ctx.Args().First(), cfg.EnvironmentAPI(region)).
			WithALBSubnets(ctx.String(options.ALBSubnets)).
			WithInstanceSubnets(ctx.String(options.InstanceSubnets)).
			WithInstanceType(ctx.String(options.InstanceType)).
			WithRegion(ctx.String(options.Region)).
			WithSize(ctx.Int(options.Size)).
			WithVPCID(ctx.String(options.VPC)), nil
	}

	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[ENVIRONMENT]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  options.VPC,
				Usage: "The vpc to create the environment in",
			},
			cli.StringFlag{
				Name:  options.ALBSubnets,
				Usage: "The subnets to place the application load balancer in",
			},
			cli.StringFlag{
				Name:  options.InstanceSubnets,
				Usage: "The subnets to place the ecs container instances in",
			},
			cli.StringFlag{
				Name:  options.Region,
				Usage: "The AWS region to create the environment in",
			},
			cli.IntFlag{
				Name:  options.Size,
				Usage: "Then number of container instances to create",
			},
			cli.StringFlag{
				Name:  options.InstanceType,
				Usage: "The type of container instances to create",
			},
			cli.StringFlag{
				Name:  options.KeyPair,
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

	options := struct {
		Force string
	}{
		Force: "force",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentDownCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithForce(ctx.Bool(options.Force))
		})
	}

	return cli.Command{
		Name:        "down",
		Usage:       "Terminates an ecso environment",
		Description: "Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  options.Force,
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
	options := struct {
		Force string
	}{
		Force: "force",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentRmCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithForce(ctx.Bool(options.Force))
		})
	}

	return cli.Command{
		Name:        "rm",
		Usage:       "Removes an ecso environment",
		Description: "Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  options.Force,
				Usage: "Required. Confirms the environment will be removed",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		DryRun string
		Force  string
	}{
		DryRun: "dry-run",
		Force:  "force",
	}

	const (
		EnvironmentUpDryRunOption = "dry-run"
		EnvironmentUpForceOption  = "force"
	)
	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeEnvironmentCommand(ctx, project, func(env *ecso.Environment) ecso.Command {
			return commands.NewEnvironmentUpCommand(env.Name, cfg.EnvironmentAPI(env.Region)).
				WithDryRun(ctx.Bool(options.DryRun)).
				WithForce(ctx.Bool(options.Force))
		})
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploys the infrastructure for an ecso environment",
		Description: "All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  options.DryRun,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
			cli.BoolFlag{
				Name:  options.Force,
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
	options := struct {
		DesiredCount string
		Route        string
		Port         string
	}{
		DesiredCount: "desired-count",
		Route:        "route",
		Port:         "port",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return commands.NewServiceAddCommand(ctx.Args().First()).
			WithDesiredCount(ctx.Int(options.DesiredCount)).
			WithRoute(ctx.String(options.Route)).
			WithPort(ctx.Int(options.Port)), nil
	}

	return cli.Command{
		Name:        "add",
		Usage:       "Adds a new service to the project",
		Description: "The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  options.DesiredCount,
				Usage: "The desired number of service instances",
			},
			cli.StringFlag{
				Name:  options.Route,
				Usage: "If set, the service will be registered with the load balancer at this route",
			},
			cli.IntFlag{
				Name:  options.Port,
				Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceDescribeCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

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
				Name:   options.Environment,
				Usage:  "The environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceDownCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
		Force       string
	}{
		Environment: "environment",
		Force:       "force",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceDownCommand(service.Name, env.Name, cfg.ServiceAPI(env.Region)).
				WithForce(ctx.Bool(options.Force))
		})
	}

	return cli.Command{
		Name:        "down",
		Usage:       "terminates a service",
		Description: "The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   options.Environment,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.BoolFlag{
				Name:  options.Force,
				Usage: "Required. Confirms the service will be terminated",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceEventsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

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
				Name:   options.Environment,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLogsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

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
				Name:   options.Environment,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		e := ctx.String(options.Environment)

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
				Name:   options.Environment,
				Usage:  "Environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServicePsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

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
				Name:   options.Environment,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceRollbackCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
		Version     string
	}{
		Environment: "environment",
		Version:     "version",
	}

	fn := func(ctx *cli.Context, cfg *config.Config) (ecso.Command, error) {
		return makeServiceCommand(ctx, project, func(service *ecso.Service, env *ecso.Environment) ecso.Command {
			return commands.NewServiceRollbackCommand(
				service.Name,
				env.Name,
				ctx.String(options.Version),
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
				Name:   options.Environment,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
			cli.StringFlag{
				Name:  options.Version,
				Usage: "The version to rollback to",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceUpCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
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
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   options.Environment,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceVersionsCliCommand(project *ecso.Project, dispatcher dispatcher.Dispatcher) cli.Command {
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

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
				Name:   options.Environment,
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
	options := struct {
		Environment string
	}{
		Environment: "environment",
	}

	name := c.Args().First()
	environmentName := c.String(options.Environment)

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
