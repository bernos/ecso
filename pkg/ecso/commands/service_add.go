package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/resources"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceAddDesiredCountOption = "desired-count"
	ServiceAddRouteOption        = "route"
	ServiceAddPortOption         = "port"
)

func NewServiceAddCommand(name string, log log.Logger) ecso.Command {
	return &serviceAddCommand{
		name:         name,
		desiredCount: 1,
		log:          log,
	}
}

type serviceAddCommand struct {
	name         string
	desiredCount int
	route        string
	port         int
	log          log.Logger
}

func (cmd *serviceAddCommand) UnmarshalCliContext(ctx *cli.Context) error {
	cmd.desiredCount = ctx.Int(ServiceAddDesiredCountOption)
	cmd.route = ctx.String(ServiceAddRouteOption)
	cmd.port = ctx.Int(ServiceAddPortOption)

	return nil
}

func (cmd *serviceAddCommand) Execute(ctx *ecso.CommandContext) error {
	project := ctx.Project

	service := &ecso.Service{
		Name:         cmd.name,
		ComposeFile:  filepath.Join("services", cmd.name, "docker-compose.yaml"),
		DesiredCount: cmd.desiredCount,
		Tags: map[string]string{
			"project": ctx.Project.Name,
			"service": cmd.name,
		},
	}

	service.SetProject(project)

	if len(cmd.route) > 0 {
		service.Route = cmd.route
		service.RoutePriority = len(ctx.Project.Services) + 1
		service.Port = cmd.port
	}

	templateData := struct {
		Service *ecso.Service
		Project *ecso.Project
	}{
		Service: service,
		Project: project,
	}

	if err := resources.WriteServiceFiles(service, templateData); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ctx.Project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(cmd.log, "Service '%s' added successfully.", cmd.name)

	cmd.log.Printf("Run `ecso service up %s --environment <ENVIRONMENT>` to deploy.\n\n", cmd.name)

	return nil
}

func (cmd *serviceAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceAddCommand) Prompt(ctx *ecso.CommandContext) error {
	var prompts = struct {
		Name         string
		DesiredCount string
		Route        string
		Port         string
	}{
		Name:         "What is the name of your service?",
		DesiredCount: "How many instances of the service would you like to run?",
		Route:        "What route would you like to expose the service at?",
		Port:         "Which container port would you like to expose?",
	}

	ui.BannerBlue(cmd.log, "Adding a new service to the %s project", ctx.Project.Name)

	if err := ui.AskStringIfEmptyVar(&cmd.name, prompts.Name, "", serviceNameValidator(ctx.Project)); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(&cmd.desiredCount, prompts.DesiredCount, 1, desiredCountValidator()); err != nil {
		return err
	}

	webChoice, err := ui.Choice("Is this a web service?", []string{"Yes", "No"})

	if err != nil {
		return err
	}

	if webChoice == 0 {
		if err := ui.AskStringIfEmptyVar(&cmd.route, prompts.Route, "/"+cmd.name, routeValidator()); err != nil {
			return err
		}

		if err := ui.AskIntIfEmptyVar(&cmd.port, prompts.Port, 80, portValidator()); err != nil {
			return err
		}
	}

	return nil
}

func serviceNameValidator(p *ecso.Project) func(string) error {
	return func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasService(val) {
			return fmt.Errorf("This project already has a service named '%s'. Please choose another name", val)
		}

		return nil
	}
}

func routeValidator() func(string) error {
	return ui.ValidateAny()
}

func desiredCountValidator() func(int) error {
	return ui.ValidateIntBetween(1, 10)
}

func portValidator() func(int) error {
	return ui.ValidateIntBetween(1, 60000)
}
