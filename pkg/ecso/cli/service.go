package cli

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/dispatcher"
	"gopkg.in/urfave/cli.v1"
)

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
