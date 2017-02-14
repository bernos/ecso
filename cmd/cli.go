package cmd

import (
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Unset string
	}{
		Unset: "unset",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewSkeletonCommand(c.Args().First(), func(opt *commands.SkeletonOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:      "TODO",
		Usage:     "TODO",
		ArgsUsage: "[TODO]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Unset,
				Usage: "TODO",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Unset string
	}{
		Unset: "unset",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		env := c.Args().First()

		return commands.NewEnvCommand(env, func(opt *commands.EnvOptions) {
			opt.Unset = c.Bool(keys.Unset)
		}), nil
	}

	return cli.Command{
		Name:      "env",
		Usage:     "Display the commands to set up the default environment for the ecso cli tool",
		ArgsUsage: "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Unset,
				Usage: "If set, output shell commands to unset all ecso environment variables",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvironmentCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	return cli.Command{
		Name:  "environment",
		Usage: "Manage ecso environments",
		Subcommands: []cli.Command{
			NewEnvironmentAddCliCommand(dispatcher),
			NewEnvironmentUpCliCommand(dispatcher),
			NewEnvironmentRmCliCommand(dispatcher),
			NewEnvironmentDescribeCliCommand(dispatcher),
			NewEnvironmentDownCliCommand(dispatcher),
		},
	}
}

func NewEnvironmentAddCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Name            string
		VPCID           string
		ALBSubnets      string
		InstanceSubnets string
		InstanceType    string
		Region          string
		Size            string
	}{
		Name:            "name",
		VPCID:           "vpc",
		ALBSubnets:      "alb-subnets",
		InstanceSubnets: "instance-subnets",
		InstanceType:    "instance-type",
		Region:          "region",
		Size:            "size",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewEnvironmentAddCommand(c.Args().First(), func(opt *commands.EnvironmentAddOptions) {
			if c.String(keys.VPCID) != "" {
				opt.VPCID = c.String(keys.VPCID)
			}

			if c.String(keys.ALBSubnets) != "" {
				opt.ALBSubnets = c.String(keys.ALBSubnets)
			}

			if c.String(keys.InstanceSubnets) != "" {
				opt.InstanceSubnets = c.String(keys.InstanceSubnets)
			}

			if c.String(keys.Region) != "" {
				opt.Region = c.String(keys.Region)
			}

			if c.Int(keys.Size) != 0 {
				opt.Size = c.Int(keys.Size)
			}

			if c.String(keys.InstanceType) != "" {
				opt.InstanceType = c.String(keys.InstanceType)
			}
		}), nil
	}

	return cli.Command{
		Name:      "add",
		Usage:     "Add a new environment to the project",
		ArgsUsage: "[ENVIRONMENT]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.VPCID,
				Usage: "The vpc to create the environment in",
			},
			cli.StringFlag{
				Name:  keys.ALBSubnets,
				Usage: "The subnets to place the application load balancer in",
			},
			cli.StringFlag{
				Name:  keys.InstanceSubnets,
				Usage: "The subnets to place the ecs container instances in",
			},
			cli.StringFlag{
				Name:  keys.Region,
				Usage: "The AWS region to create the environment in",
			},
			cli.IntFlag{
				Name:  keys.Size,
				Usage: "Then number of container instances to create",
			},
			cli.StringFlag{
				Name:  keys.InstanceType,
				Usage: "The type of container instances to create",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvironmentDescribeCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")

		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		return commands.NewEnvironmentDescribeCommand(env, func(opt *commands.EnvironmentDescribeOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:      "describe",
		Usage:     "Describes an ecso environment",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvironmentDownCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	keys := struct {
		Force string
	}{
		Force: "force",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		force := c.Bool(keys.Force)
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")
		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		if !force {
			return nil, NewOptionRequiredError(keys.Force)
		}

		return commands.NewEnvironmentDownCommand(env, func(opt *commands.EnvironmentDownOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:        "down",
		Usage:       "Terminates an ecso environment",
		Description: "Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Force,
				Usage: "Required. Confirms the environment will be stopped",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvironmentRmCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	keys := struct {
		Force string
	}{
		Force: "force",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		force := c.Bool(keys.Force)
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")
		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		if !force {
			return nil, NewOptionRequiredError(keys.Force)
		}

		return commands.NewEnvironmentRmCommand(env, func(opt *commands.EnvironmentRmOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:        "rm",
		Usage:       "Removes an ecso environment",
		Description: "Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.Force,
				Usage: "Required. Confirms the environment will be removed",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewEnvironmentUpCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		DryRun string
	}{
		DryRun: "dry-run",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")
		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		return commands.NewEnvironmentUpCommand(env, func(opt *commands.EnvironmentUpOptions) {
			opt.DryRun = c.Bool(keys.DryRun)
		}), nil
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploys the infrastructure for an ecso environment",
		Description: "All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.",
		ArgsUsage:   "ENVIRONMENT",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  keys.DryRun,
				Usage: "If set, list pending changes, but do not execute the updates.",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewInitCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewInitCommand(c.Args().First()), nil
	}

	return cli.Command{
		Name:        "init",
		Usage:       "Initialise a new ecso project",
		Description: "Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.",
		ArgsUsage:   "[PROJECT]",
		Action:      MakeAction(dispatcher, fromCliContext, ecso.SkipEnsureProjectExists()),
	}
}

func NewServiceCliCommand(dispatcher ecso.Dispatcher) cli.Command {
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
		},
	}
}

func NewServiceAddCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		DesiredCount string
		Route        string
		Port         string
	}{
		DesiredCount: "desired-count",
		Route:        "route",
		Port:         "port",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewServiceAddCommand(c.Args().First(), func(opt *commands.ServiceAddOptions) {
			opt.DesiredCount = c.Int(keys.DesiredCount)
			opt.Route = c.String(keys.Route)
			opt.Port = c.Int(keys.Port)
		}), nil
	}

	return cli.Command{
		Name:        "add",
		Usage:       "Adds a new service to the project",
		Description: "The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  keys.DesiredCount,
				Usage: "The desired number of service instances",
			},
			cli.StringFlag{
				Name:  keys.Route,
				Usage: "If set, the service will be registered with the load balancer at this route",
			},
			cli.IntFlag{
				Name:  keys.Port,
				Usage: "If set, the loadbalancer will bind to this port of the web container in this service",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceDescribeCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		service := c.Args().First()
		env := c.String(keys.Environment)

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServiceDescribeCommand(service, env, func(opt *commands.ServiceDescribeOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:        "describe",
		Usage:       "Lists details of a deployed service",
		Description: "Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceDownCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
		Force       string
	}{
		Environment: "environment",
		Force:       "force",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		force := c.Bool(keys.Force)
		service := c.Args().First()
		env := c.String(keys.Environment)

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		if !force {
			return nil, NewOptionRequiredError(keys.Force)
		}

		return commands.NewServiceDownCommand(service, env, func(opt *commands.ServiceDownOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:        "down",
		Usage:       "terminates a service",
		Description: "The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceEventsCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Name        string
		Environment string
	}{
		Name:        "name",
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		name := c.String(keys.Name)
		env := c.String(keys.Environment)

		if name == "" {
			return nil, NewOptionRequiredError(keys.Name)
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServiceEventsCommand(name, env, func(opt *commands.ServiceEventsOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:      "events",
		Usage:     "List ECS events for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  keys.Name,
				Usage: "The service to list events for",
			},
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceLogsCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		service := c.Args().First()
		env := c.String(keys.Environment)

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServiceLogsCommand(service, env, func(opt *commands.ServiceLogsOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:      "logs",
		Usage:     "output service logs",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceLsCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		env := c.String(keys.Environment)

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServiceLsCommand(env, func(opt *commands.ServiceLsOptions) {
			// TODO: populate options from c
		}), nil
	}

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
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServicePsCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		name := c.Args().First()
		env := c.String(keys.Environment)

		if name == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServicePsCommand(name, env, func(opt *commands.ServicePsOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:      "ps",
		Usage:     "Show running tasks for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}

func NewServiceUpCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Environment string
	}{
		Environment: "environment",
	}

	fromCliContext := func(c *cli.Context) (ecso.Command, error) {
		name := c.Args().First()
		env := c.String(keys.Environment)

		if name == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if env == "" {
			return nil, NewOptionRequiredError(keys.Environment)
		}

		return commands.NewServiceUpCommand(name, env, func(opt *commands.ServiceUpOptions) {
			// TODO: populate options from c
		}), nil
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploy a service",
		Description: "The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   keys.Environment,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fromCliContext),
	}
}
