package addservice

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type Options struct {
	Name         string
	DesiredCount int
	Route        string
	Port         int
}

type command struct {
	options *Options
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger
		project = ctx.Project
	)

	if project.HasService(cmd.options.Name) {
		return fmt.Errorf("Service '%s' already exists", cmd.options.Name)
	}

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

	if err := writeFiles(project, service); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ctx.Project.Save(); err != nil {
		return err
	}

	log.BannerGreen("Service '%s' added successfully. Now run `ecso service up --name %s --environment <environment>` to deploy", cmd.options.Name, cmd.options.Name)

	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	return nil
}

func New(name string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:         name,
		DesiredCount: 1,
	}

	for _, option := range options {
		option(o)
	}

	return &command{
		options: o,
	}
}

func writeFiles(project *ecso.Project, service *ecso.Service) error {
	var (
		composeFile        = filepath.Join(project.Dir(), service.ComposeFile)
		cloudFormationFile = filepath.Join(project.Dir(), ".ecso/services", service.Name, "stack.yaml")
		templateData       = struct {
			Service *ecso.Service
			Project *ecso.Project
		}{
			Service: service,
			Project: project,
		}
	)

	var composeFileTemplate *template.Template
	var cloudFormationTemplate *template.Template

	if len(service.Route) > 0 {
		composeFileTemplate = webServiceComposeFileTemplate
		cloudFormationTemplate = webServiceCloudFormationTemplate
	} else {
		composeFileTemplate = workerComposeFileTemplate
		cloudFormationTemplate = workerCloudFormationTemplate
	}

	if err := util.WriteFileFromTemplate(composeFile, composeFileTemplate, templateData); err != nil {
		return err
	}

	if err := util.WriteFileFromTemplate(cloudFormationFile, cloudFormationTemplate, templateData); err != nil {
		return err
	}

	return nil
}
