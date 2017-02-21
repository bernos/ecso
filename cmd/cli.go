package cmd

import (
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/commands"
	"gopkg.in/urfave/cli.v1"
)

func CliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewSkeletonCommand(c.Args().First()), nil
	}

	return cli.Command{
		Name:      "TODO",
		Usage:     "TODO",
		ArgsUsage: "[TODO]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  commands.EnvUnsetOption,
				Usage: "TODO",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewEnvCommand(c.Args().First()), nil
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
	fn := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewEnvironmentAddCommand(c.Args().First()), nil
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
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentDescribeCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")

		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		return commands.NewEnvironmentDescribeCommand(env), nil
	}

	return cli.Command{
		Name:      "describe",
		Usage:     "Describes an ecso environment",
		ArgsUsage: "ENVIRONMENT",
		Action:    MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentDownCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	keys := struct {
		Force string
	}{
		Force: "force",
	}

	fn := func(c *cli.Context) (ecso.Command, error) {
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

		return commands.NewEnvironmentDownCommand(env), nil
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
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentRmCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	keys := struct {
		Force string
	}{
		Force: "force",
	}

	fn := func(c *cli.Context) (ecso.Command, error) {
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

		return commands.NewEnvironmentRmCommand(env), nil
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
		Action: MakeAction(dispatcher, fn),
	}
}

func NewEnvironmentUpCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		env := c.Args().First()

		if env == "" {
			env = os.Getenv("ECSO_ENVIRONMENT")
		}

		if env == "" {
			return nil, NewArgumentRequiredError("environment")
		}

		return commands.NewEnvironmentUpCommand(env), nil
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
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewInitCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewInitCommand(c.Args().First()), nil
	}

	return cli.Command{
		Name:        "init",
		Usage:       "Initialise a new ecso project",
		Description: "Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.",
		ArgsUsage:   "[PROJECT]",
		Action:      MakeAction(dispatcher, fn, ecso.SkipEnsureProjectExists()),
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
	fn := func(c *cli.Context) (ecso.Command, error) {
		return commands.NewServiceAddCommand(c.Args().First()), nil
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

func NewServiceDescribeCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	fn := func(c *cli.Context) (ecso.Command, error) {
		service := c.Args().First()

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		return commands.NewServiceDescribeCommand(service), nil
	}

	return cli.Command{
		Name:        "describe",
		Usage:       "Lists details of a deployed service",
		Description: "Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceDescribeEnvironmentOption,
				Usage:  "The environment to query",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceDownCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	keys := struct {
		Force string
	}{
		Force: "force",
	}

	fn := func(c *cli.Context) (ecso.Command, error) {
		force := c.Bool(keys.Force)
		service := c.Args().First()

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		if !force {
			return nil, NewOptionRequiredError(keys.Force)
		}

		return commands.NewServiceDownCommand(service), nil
	}

	return cli.Command{
		Name:        "down",
		Usage:       "terminates a service",
		Description: "The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceDownEnvironmentOption,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceEventsCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	fn := func(c *cli.Context) (ecso.Command, error) {
		service := c.Args().First()

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		return commands.NewServiceEventsCommand(service), nil
	}

	return cli.Command{
		Name:      "events",
		Usage:     "List ECS events for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceEventsEnvironmentOption,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLogsCliCommand(dispatcher ecso.Dispatcher) cli.Command {
	fn := func(c *cli.Context) (ecso.Command, error) {
		service := c.Args().First()

		if service == "" {
			return nil, NewArgumentRequiredError("service")
		}

		return commands.NewServiceLogsCommand(service), nil
	}

	return cli.Command{
		Name:      "logs",
		Usage:     "output service logs",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceLogsEnvironmentOption,
				Usage:  "The environment to terminate the service from",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceLsCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	fn := func(c *cli.Context) (ecso.Command, error) {
		env := c.String(commands.ServiceLsEnvironmentOption)

		if env == "" {
			return nil, NewOptionRequiredError(commands.ServiceLsEnvironmentOption)
		}

		return commands.NewServiceLsCommand(env), nil
	}

	return cli.Command{
		Name:  "ls",
		Usage: "List services",
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

func NewServicePsCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	fn := func(c *cli.Context) (ecso.Command, error) {
		name := c.Args().First()

		if name == "" {
			return nil, NewArgumentRequiredError("service")
		}

		return commands.NewServicePsCommand(name), nil
	}

	return cli.Command{
		Name:      "ps",
		Usage:     "Show running tasks for a service",
		ArgsUsage: "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServicePsEnvironmentOption,
				Usage:  "The name of the environment",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}

func NewServiceUpCliCommand(dispatcher ecso.Dispatcher) cli.Command {

	fn := func(c *cli.Context) (ecso.Command, error) {
		name := c.Args().First()

		if name == "" {
			return nil, NewArgumentRequiredError("service")
		}

		return commands.NewServiceUpCommand(name), nil
	}

	return cli.Command{
		Name:        "up",
		Usage:       "Deploy a service",
		Description: "The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.",
		ArgsUsage:   "SERVICE",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   commands.ServiceUpEnvironmentOption,
				Usage:  "The name of the environment to deploy to",
				EnvVar: "ECSO_ENVIRONMENT",
			},
		},
		Action: MakeAction(dispatcher, fn),
	}
}
