package resources

//go:generate go-bindata -pkg $GOPACKAGE -o resources-generated.go ./...

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
)

var (
	// Cloudformation template for our default web service
	WebServiceCloudFormationTemplate = NewTextFile(MustParseTemplateAsset("stack.yaml", "services/web/cloudformation/stack.yaml"))

	// Docker compose file for default web service
	WebServiceDockerComposeFile = NewTextFile(MustParseTemplateAsset("docker-compose.yaml", "services/web/docker-compose.yaml"))

	// Cloudformation template for default worker service
	WorkerServiceCloudFormationTemplate = NewTextFile(MustParseTemplateAsset("stack.yaml", "services/worker/cloudformation/stack.yaml"))

	// Docker compose file for default worker service
	WorkerServiceDockerComposeFile = NewTextFile(MustParseTemplateAsset("docker-compose.yaml", "services/worker/docker-compose.yaml"))

	// EnvironmentFiles is the list of all files needed to build an environment
	EnvironmentFiles = environmentFiles()

	InstanceDrainerLambdaVersion = "1.0.0"

	ServiceDiscoveryLambdaVersion = "1.0.0"
)

func environmentFiles() []Resource {
	files := environmentCfnTemplates()
	files = append(files, environmentLambdas()...)

	return files
}

func environmentLambdas() []Resource {
	zip := func(name, version string) string {
		return fmt.Sprintf("%s/lambda/%s-%s.zip", ecso.EnvironmentResourceDir, name, version)
	}

	src := func(lambda, path string) string {
		return filepath.Join("environment", "lambda", lambda, path)
	}

	return []Resource{
		NewZipFile(zip("instance-drainer", InstanceDrainerLambdaVersion),
			MustParseTemplateAsset("index.py", src("instance-drainer", "index.py"))),

		NewZipFile(zip("service-discovery", ServiceDiscoveryLambdaVersion),
			MustParseTemplateAsset("index.js", src("service-discovery", "index.js"))),
	}
}

func environmentCfnTemplates() []Resource {
	cfnTemplates := []string{
		"alarms.yaml",
		"dd-agent.yaml",
		"dns-cleaner.yaml",
		"ecs-cluster.yaml",
		"instance-drainer.yaml",
		"load-balancers.yaml",
		"logging.yaml",
		"security-groups.yaml",
		"service-discovery.yaml",
		"sns.yaml",
		"stack.yaml",
	}

	src := func(path string) string {
		return filepath.Join("environment", "cloudformation", path)
	}

	dst := func(path string) string {
		return filepath.Join(ecso.EnvironmentCloudFormationDir, path)
	}

	files := make([]Resource, 0)

	for _, cfnTemplate := range cfnTemplates {
		files = append(files, NewTextFile(MustParseTemplateAsset(dst(cfnTemplate), src(cfnTemplate))))
	}

	return files
}

type Resource interface {
	Filename() string
	WriteTo(w io.Writer, data interface{}) error
}

func MustParseTemplate(name, body string) *template.Template {
	return template.Must(template.New(name).Parse(body))
}

func MustParseTemplateAsset(name, assetPath string) *template.Template {
	a := MustAsset(assetPath)
	t := template.Must(template.New(name).Parse(string(a)))

	return t
}
