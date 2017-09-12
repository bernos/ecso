package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/resources"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

const (
	ServiceAddDesiredCountOption = "desired-count"
	ServiceAddRouteOption        = "route"
	ServiceAddPortOption         = "port"
)

func NewServiceAddCommand(name string) *ServiceAddCommand {
	return &ServiceAddCommand{
		name:         name,
		desiredCount: 1,
	}
}

type ServiceAddCommand struct {
	name         string
	desiredCount int
	route        string
	port         int
}

func (cmd *ServiceAddCommand) WithDesiredCount(x int) *ServiceAddCommand {
	cmd.desiredCount = x
	return cmd
}

func (cmd *ServiceAddCommand) WithRoute(route string) *ServiceAddCommand {
	cmd.route = route
	return cmd
}

func (cmd *ServiceAddCommand) WithPort(x int) *ServiceAddCommand {
	cmd.port = x
	return cmd
}

func (cmd *ServiceAddCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	project := ctx.Project
	green := ui.NewBannerWriter(w, ui.GreenBold)

	if err := cmd.prompt(ctx, r, w); err != nil {
		return err
	}

	service := &ecso.Service{
		Name:         cmd.name,
		ComposeFile:  filepath.Join("services", cmd.name, "docker-compose.yaml"),
		DesiredCount: cmd.desiredCount,
		Tags: map[string]string{
			"project": ctx.Project.Name,
			"service": cmd.name,
		},
	}

	if len(cmd.route) > 0 {
		service.Route = cmd.route
		service.RoutePriority = len(ctx.Project.Services) + 1
		service.Port = cmd.port
	}

	project.AddService(service)

	if err := cmd.createResources(project, service); err != nil {
		return err
	}

	if err := ctx.Project.Save(); err != nil {
		return err
	}

	fmt.Fprintf(green, "Service '%s' added successfully.", cmd.name)
	fmt.Fprintf(w, "Run `ecso service up %s --environment <ENVIRONMENT>` to deploy.\n\n", cmd.name)

	return nil
}

func (cmd *ServiceAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *ServiceAddCommand) createResources(project *ecso.Project, service *ecso.Service) error {
	cfnWriter := resources.NewFileSystemResourceWriter(service.GetCloudFormationTemplateDir())
	resourceWriter := resources.NewFileSystemResourceWriter(service.Dir())

	templateData := struct {
		Service *ecso.Service
		Project *ecso.Project
	}{
		Service: service,
		Project: project,
	}

	var serviceResources *resources.ServiceResources

	if len(service.Route) > 0 {
		serviceResources = &resources.WebService
	} else {
		serviceResources = &resources.WorkerService
	}

	if err := cfnWriter.WriteResource(serviceResources.CloudFormationTemplate, templateData); err != nil {
		return err
	}

	if err := resourceWriter.WriteResource(serviceResources.ComposeFile, templateData); err != nil {
		return err
	}

	return nil
}

func (cmd *ServiceAddCommand) prompt(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	blue := ui.NewBannerWriter(w, ui.BlueBold)

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

	fmt.Fprintf(blue, "Adding a new service to the %s project", ctx.Project.Name)

	if err := ui.AskStringIfEmptyVar(r, w, &cmd.name, prompts.Name, "", serviceNameValidator(ctx.Project)); err != nil {
		return err
	}

	if err := ui.AskIntIfEmptyVar(r, w, &cmd.desiredCount, prompts.DesiredCount, 1, desiredCountValidator()); err != nil {
		return err
	}

	webChoice, err := ui.Choice(r, w, "Is this a web service?", []string{"Yes", "No"})

	if err != nil {
		return err
	}

	if webChoice == 0 {
		if err := ui.AskStringIfEmptyVar(r, w, &cmd.route, prompts.Route, "/"+cmd.name, routeValidator()); err != nil {
			return err
		}

		if err := ui.AskIntIfEmptyVar(r, w, &cmd.port, prompts.Port, 80, portValidator()); err != nil {
			return err
		}
	}

	return nil
}

func serviceNameValidator(p *ecso.Project) ui.StringValidator {
	return ui.StringValidatorFunc(func(val string) error {
		if val == "" {
			return fmt.Errorf("Name is required")
		}

		if p.HasService(val) {
			return fmt.Errorf("This project already has a service named '%s'. Please choose another name", val)
		}

		return nil
	})
}

func routeValidator() ui.StringValidator {
	return ui.ValidateAny()
}

func desiredCountValidator() ui.IntValidator {
	return ui.ValidateIntBetween(1, 10)
}

func portValidator() ui.IntValidator {
	return ui.ValidateIntBetween(1, 60000)
}
