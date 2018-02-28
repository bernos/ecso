package cli

import (
	"fmt"
	"os"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

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
