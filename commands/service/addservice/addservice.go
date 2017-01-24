package addservice

import (
	"fmt"
	"path/filepath"

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
		log = ctx.Config.Logger
	)

	projectDir, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return err
	}

	if err := promptForMissingOptions(cmd.options, ctx); err != nil {
		return err
	}

	if err := validateOptions(cmd.options); err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[cmd.options.Name]; ok {
		return fmt.Errorf("Service '%s' already exists", cmd.options.Name)
	}

	log.BannerBlue("Adding '%s' service", cmd.options.Name)

	service := ecso.Service{
		Name:          cmd.options.Name,
		ComposeFile:   filepath.Join("services", cmd.options.Name, "docker-compose.yaml"),
		DesiredCount:  cmd.options.DesiredCount,
		Route:         cmd.options.Route,
		RoutePriority: len(ctx.Project.Services) + 1,
		Port:          cmd.options.Port,
		Tags: map[string]string{
			"project": ctx.Project.Name,
		},
	}

	if err := writeFiles(projectDir, service); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ecso.SaveCurrentProject(ctx.Project); err != nil {
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

func promptForMissingOptions(options *Options, ctx *ecso.CommandContext) error {
	// TODO prompt for missing options
	return nil
}

func writeFiles(projectDir string, service ecso.Service) error {
	var (
		composeFile        = filepath.Join(projectDir, service.ComposeFile)
		cloudFormationFile = filepath.Join(projectDir, ".ecso/services", service.Name, "stack.yaml")
		templateData       = struct {
			Service ecso.Service
		}{
			Service: service,
		}
	)

	if err := util.WriteFileFromTemplate(composeFile, composeFileTemplate, templateData); err != nil {
		return err
	}

	if err := util.WriteFileFromTemplate(cloudFormationFile, cloudFormationTemplate, templateData); err != nil {
		return err
	}

	return nil
}

func validateOptions(opt *Options) error {
	if opt.Name == "" {
		return fmt.Errorf("Name is required")
	}
	return nil
}
