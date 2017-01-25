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

	if err := promptForMissingOptions(cmd.options, ctx); err != nil {
		return err
	}

	if project.HasService(cmd.options.Name) {
		return fmt.Errorf("Service '%s' already exists", cmd.options.Name)
	}

	log.BannerBlue("Adding '%s' service", cmd.options.Name)

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

	if err := writeFiles(project.Dir(), service); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ctx.Project.Save(); err != nil {
		return err
	}

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

func writeFiles(projectDir string, service *ecso.Service) error {
	var (
		composeFile        = filepath.Join(projectDir, service.ComposeFile)
		cloudFormationFile = filepath.Join(projectDir, ".ecso/services", service.Name, "stack.yaml")
		templateData       = struct {
			Service *ecso.Service
		}{
			Service: service,
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
