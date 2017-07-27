package resources

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
)

var (
	// Cloudformation template for our default web service
	WebServiceCloudFormationTemplate = NewTextFile(MustParseTemplate("stack.yaml", webServiceCloudFormationTemplate))

	// Docker compose file for default web service
	WebServiceDockerComposeFile = NewTextFile(MustParseTemplate("docker-compose.yaml", webServiceComposeFileTemplate))

	// Cloudformation template for default worker service
	WorkerServiceCloudFormationTemplate = NewTextFile(MustParseTemplate("stack.yaml", workerCloudFormationTemplate))

	// Docker compose file for default worker service
	WorkerServiceDockerComposeFile = NewTextFile(MustParseTemplate("docker-compose.yaml", workerComposeFileTemplate))

	// EnvironmentFiles is the list of all files needed to build an environment
	EnvironmentFiles = &Resources{[]Resource{
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/alarms.yaml", environmentAlarmsTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/dd-agent.yaml", environmentDataDogTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/dns-cleaner.yaml", environmentDNSCleanerTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/ecs-cluster.yaml", environmentClusterTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/instance-drainer.yaml", environmentInstanceDrainerLambda)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/load-balancers.yaml", environmentALBTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/logging.yaml", environmentLoggingTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/security-groups.yaml", environmentSecurityGroupTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/sns.yaml", environmentSNSTemplate)),
		NewTextFile(MustParseTemplate(ecso.EnvironmentCloudFormationDir+"/stack.yaml", environmentStackTemplate)),

		NewZipFile(fmt.Sprintf("%s/lambda/instance-drainer-%s.zip", ecso.EnvironmentResourceDir, InstanceDrainerLambdaVersion),
			template.Must(template.New("index.py").Parse(environmentInstanceDrainerLambdaSource))),
	}}
)

func MustParseTemplate(name, body string) *template.Template {
	return template.Must(template.New(name).Parse(body))
}

func WriteServiceFiles(s *ecso.Service, data interface{}) error {
	if len(s.Route) > 0 {
		if err := WebServiceCloudFormationTemplate.WriteTo(s.GetCloudFormationTemplateDir(), data); err != nil {
			return err
		}

		if err := WebServiceDockerComposeFile.WriteTo(s.Dir(), data); err != nil {
			return err
		}
	} else {
		if err := WorkerServiceCloudFormationTemplate.WriteTo(s.GetCloudFormationTemplateDir(), data); err != nil {
			return err
		}

		if err := WorkerServiceDockerComposeFile.WriteTo(s.Dir(), data); err != nil {
			return err
		}
	}

	return nil
}

type Resources struct {
	rs []Resource
}

func (rs *Resources) Add(r Resource) {
	rs.rs = append(rs.rs, r)
}

func (rs *Resources) WriteTo(basePath string, data interface{}) error {
	for _, r := range rs.rs {
		if err := r.WriteTo(basePath, data); err != nil {
			return err
		}
	}
	return nil
}

type Resource interface {
	WriteTo(basePath string, data interface{}) error
}

type textFile struct {
	*template.Template
}

func NewTextFile(tmpl *template.Template) Resource {
	return &textFile{tmpl}
}

func (r *textFile) WriteTo(basePath string, data interface{}) error {
	return writeTemplateTo(filepath.Join(basePath, r.Name()), r.Template, data)
}

func writeTemplateTo(filename string, tmpl *template.Template, data interface{}) error {
	w, err := getWriter(filename)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func getWriter(filename string) (io.Writer, error) {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(filename)
}

func NewZipFile(file string, entries ...*template.Template) Resource {
	return &zipFile{
		file:    file,
		entries: entries,
	}
}

type zipFile struct {
	file    string
	entries []*template.Template
}

func (z *zipFile) WriteTo(basePath string, data interface{}) error {
	w, err := getWriter(filepath.Join(basePath, z.file))
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)

	for _, tmpl := range z.entries {
		f, err := zipWriter.Create(tmpl.Name())
		if err != nil {
			return err
		}

		if err := tmpl.Execute(f, data); err != nil {
			return err
		}
	}

	return zipWriter.Close()
}
