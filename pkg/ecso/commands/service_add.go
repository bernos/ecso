package commands

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/resources"
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

	ctx.Project.AddService(service)

	if err := cmd.createResources(ctx.Project, service); err != nil {
		return err
	}

	return ctx.Project.Save()
}

func (cmd *ServiceAddCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.name == "" {
		return fmt.Errorf("Service name required")
	}

	if cmd.desiredCount == 0 {
		return fmt.Errorf("Desired count required")
	}

	if cmd.route != "" && cmd.port == 0 {
		return fmt.Errorf("Port is required")
	}
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
