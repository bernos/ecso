package addservice

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
)

var (
	composeFileTemplate    = template.Must(template.New("composeFile").Parse(composeFileTemplateString))
	cloudFormationTemplate = template.Must(template.New("cloudFormationTemplate").Parse(cloudFormationTemplateString))
)

type Options struct {
	Name         string
	DesiredCount int
	Route        string
	Port         int
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

	// TODO prompt for missing options

	if err := validateOptions(cmd.options); err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[cmd.options.Name]; ok {
		return fmt.Errorf("Service '%s' already exists", cmd.options.Name)
	}

	log.BannerBlue("Adding '%s' service", cmd.options.Name)

	service := ecso.Service{
		Name:         cmd.options.Name,
		ComposeFile:  filepath.Join("services", cmd.options.Name, "docker-compose.yaml"),
		DesiredCount: cmd.options.DesiredCount,
		Route:        cmd.options.Route,
		Port:         cmd.options.Port,
		Tags: map[string]string{
			"project": ctx.Project.Name,
		},
	}

	composeFile := filepath.Join(projectDir, service.ComposeFile)
	cloudFormationFile := filepath.Join(projectDir, ".ecso/services", cmd.options.Name, "resources.yaml")
	templateData := struct {
		Service ecso.Service
	}{
		Service: service,
	}

	if err := writeFileFromTemplate(composeFile, composeFileTemplate, templateData); err != nil {
		return err
	}

	if err := writeFileFromTemplate(cloudFormationFile, cloudFormationTemplate, templateData); err != nil {
		return err
	}

	ctx.Project.AddService(service)

	if err := ecso.SaveCurrentProject(ctx.Project); err != nil {
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

func writeFileFromTemplate(filename string, tmpl *template.Template, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	w, err := os.Create(filename)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}
