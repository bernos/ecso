package templates

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
)

func GetEnvironmentTemplates(project *ecso.Project, env *ecso.Environment) map[string]*template.Template {
	p := func(file string) string {
		return filepath.Join(env.GetCloudFormationTemplateDir(), file)
	}

	return map[string]*template.Template{
		p("stack.yaml"):           environmentStackTemplate,
		p("logging.yaml"):         environmentLoggingTemplate,
		p("ecs-cluster.yaml"):     environmentClusterTemplate,
		p("load-balancers.yaml"):  environmentALBTemplate,
		p("security-groups.yaml"): environmentSecurityGroupTemplate,
		p("dd-agent.yaml"):        environmentDataDogTemplate,
		p("sns.yaml"):             environmentSNSTemplate,
		p("alarms.yaml"):          environmentAlarmsTemplate,
		p("dns-cleaner.yaml"):     environmentDNSCleanerTemplate,
	}
}

func GetServiceTemplates(project *ecso.Project, service *ecso.Service) map[string]*template.Template {
	compose := filepath.Join(project.Dir(), service.ComposeFile)
	cfn := filepath.Join(project.Dir(), ".ecso/services", service.Name, "stack.yaml")

	if len(service.Route) > 0 {
		return map[string]*template.Template{
			compose: webServiceComposeFileTemplate,
			cfn:     webServiceCloudFormationTemplate,
		}
	}

	return map[string]*template.Template{
		compose: workerComposeFileTemplate,
		cfn:     workerCloudFormationTemplate,
	}
}

// WriteFile renders a template with `data` and write the result to a file
func WriteFile(filename string, tmpl *template.Template, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	w, err := os.Create(filename)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

// WriteFiles renders multiple templates to multiple files. `templateMap` is a
// map of filename:template. Each template in the map is processed with `data`
func WriteFiles(templateMap map[string]*template.Template, data interface{}) error {
	for file, tmpl := range templateMap {
		if err := WriteFile(file, tmpl, data); err != nil {
			return err
		}
	}
	return nil
}

// WriteEnvironmentFiles renders and writes project template files to disk
func WriteEnvironmentFiles(project *ecso.Project, env *ecso.Environment, data interface{}) error {
	return WriteFiles(GetEnvironmentTemplates(project, env), data)
}

// WriteServiceFiles renders and writes service template files to disk
func WriteServiceFiles(project *ecso.Project, service *ecso.Service, data interface{}) error {
	return WriteFiles(GetServiceTemplates(project, service), data)
}
