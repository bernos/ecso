package commands

import (
	"fmt"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/templates"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type ServiceAddOptions struct {
	Name         string
	DesiredCount int
	Route        string
	Port         int
}

func NewServiceAddCommand(name string, options ...func(*ServiceAddOptions)) ecso.Command {
	o := &ServiceAddOptions{
		Name:         name,
		DesiredCount: 1,
	}

	for _, option := range options {
		option(o)
	}

	return &serviceAddCommand{
		options: o,
	}
}

type serviceAddCommand struct {
	options *ServiceAddOptions
}

func (cmd *serviceAddCommand) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger()
		project = ctx.Project
	)

	service := &ecso.Service{
		Name:         cmd.options.Name,
		ComposeFile:  filepath.Join("services", cmd.options.Name, "docker-compose.yaml"),
		DesiredCount: cmd.options.DesiredCount,
		Tags: map[string]string{
			"project": ctx.Project.Name,
			"service": cmd.options.Name,
		},
	}

	if len(cmd.options.Route) > 0 {
		service.Route = cmd.options.Route
		service.RoutePriority = len(ctx.Project.Services) + 1
		service.Port = cmd.options.Port
	}

	templateData := struct {
		Service *ecso.Service
		Project *ecso.Project
	}{
		Service: service,
		Project: project,
	}

	if err := templates.WriteServiceFiles(project, service, templateData); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ctx.Project.Save(); err != nil {
		return err
	}

	ui.BannerGreen(log, "Service '%s' added successfully. Now run `ecso service up --name %s --environment <environment>` to deploy", cmd.options.Name, cmd.options.Name)

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

	opt := cmd.options

	if err := ui.AskStringIfEmptyVar(&opt.Name, prompts.Name, "", serviceNameValidator(ctx.Project)); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(&opt.DesiredCount, prompts.DesiredCount, 1, desiredCountValidator()); err != nil {
		return err
	}

	webChoice, err := ui.Choice("Is this a web service?", []string{"Yes", "No"})

	if err != nil {
		return err
	}

	if webChoice == 0 {
		if err := ui.AskStringIfEmptyVar(&opt.Route, prompts.Route, "/"+opt.Name, routeValidator()); err != nil {
			return err
		}

		if err := ui.AskIntIfEmptyVar(&opt.Port, prompts.Port, 80, portValidator()); err != nil {
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
