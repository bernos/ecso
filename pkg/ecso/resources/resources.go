package resources

import (
	"fmt"
	"io"
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
	EnvironmentFiles = []Resource{
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
	}
)

type Resource interface {
	Filename() string
	WriteTo(w io.Writer, data interface{}) error
}

func MustParseTemplate(name, body string) *template.Template {
	return template.Must(template.New(name).Parse(body))
}
