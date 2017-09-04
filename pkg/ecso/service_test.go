package ecso

import (
	"fmt"
	"reflect"
	"testing"

	lconfig "github.com/docker/libcompose/config"
)

func makeTestProject() *Project {
	return NewProject("/project", "my-project", "1")
}

func makeTestEnvironment() *Environment {
	return &Environment{
		project: makeTestProject(),
		Name:    "test",
	}
}

func makeTestService() *Service {
	return &Service{
		project:     makeTestProject(),
		Name:        "my-service",
		ComposeFile: "testdata/services/my-service/docker-compose.yaml",
	}
}

func assertEqual(want, got interface{}, t *testing.T) {
	if want != got {
		t.Errorf("Wanted %s, got %s", want, got)
	}
}

func TestDir(t *testing.T) {
	assertEqual("/project/services/my-service",
		makeTestService().Dir(), t)
}

func TestGetCloudFormationTemplateDir(t *testing.T) {
	assertEqual("/project/.ecso/services/my-service",
		makeTestService().GetCloudFormationTemplateDir(), t)
}

func TestGetCloudFormationTemplateFile(t *testing.T) {
	assertEqual("/project/.ecso/services/my-service/stack.yaml",
		makeTestService().GetCloudFormationTemplateFile(), t)
}

func TestGetCloudFormationStackName(t *testing.T) {
	env := &Environment{Name: "test"}

	assertEqual("my-project-test-my-service",
		makeTestService().GetCloudFormationStackName(env), t)
}

func TestGetDeploymentBucketPrefixForVersion(t *testing.T) {
	env := makeTestEnvironment()
	version := "1.0.0"

	assertEqual("my-project-test/services/my-service/1.0.0",
		makeTestService().GetDeploymentBucketPrefixForVersion(env, version), t)
}

func TestGetDeploymentBucket(t *testing.T) {
	env := makeTestEnvironment()

	assertEqual("my-project-test/services/my-service",
		makeTestService().GetDeploymentBucketPrefix(env), t)
}

func TestGetCloudWatchLogGroup(t *testing.T) {
	env := makeTestEnvironment()

	assertEqual("my-project-test",
		makeTestService().GetCloudWatchLogGroup(env), t)
}

func TestGetCloudWatchLogStreamPrefix(t *testing.T) {
	env := makeTestEnvironment()

	assertEqual("services/my-service",
		makeTestService().GetCloudWatchLogStreamPrefix(env), t)
}

func TestGetECSTaskDefinitionName(t *testing.T) {
	env := makeTestEnvironment()

	assertEqual("my-project-test-my-service",
		makeTestService().GetECSTaskDefinitionName(env), t)
}

func TestGetEnvFile(t *testing.T) {
	env := makeTestEnvironment()

	assertEqual("testdata/services/my-service/.test.env",
		makeTestService().GetEnvFile(env), t)
}

func TestGetEnvironmentLookup(t *testing.T) {
	env := makeTestEnvironment()
	service := makeTestService()

	service.Environments = map[string]ServiceConfiguration{
		env.Name: ServiceConfiguration{
			Env: map[string]string{
				"BOTH":     "value-from-ecso-project-file",
				"ECSO_ENV": "ecso-env-value",
			},
		},
	}

	lookup, err := service.GetEnvironmentLookup(env)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input  string
		output []string
	}{
		{
			input:  "ECSO_ENV",
			output: []string{"ECSO_ENV=ecso-env-value"},
		},
		{
			input:  "ENVFILE_VAR",
			output: []string{"ENVFILE_VAR=hello"},
		},
		{
			input:  "BOTH",
			output: []string{"BOTH=value-from-envfile"},
		},
		{
			input:  "ECSO_ENVIRONMENT",
			output: []string{"ECSO_ENVIRONMENT=test"},
		},
	}

	for _, test := range tests {
		got := lookup.Lookup(test.input, nil)

		if !reflect.DeepEqual(test.output, got) {
			t.Errorf("Wanted %q, got %q", test.output, got)
		}
	}
}

func TestSetProject(t *testing.T) {
	project := makeTestProject()
	service := makeTestService()
	service.SetProject(project)

	if !reflect.DeepEqual(project, service.project) {
		t.Error("Expected projects to be equal")
	}
}

func TestGetECSTaskDefinition(t *testing.T) {
	service := makeTestService()
	env := makeTestEnvironment()
	td, err := service.GetECSTaskDefinition(env)
	if err != nil {
		t.Fatal(err)
	}

	if len(td.ContainerDefinitions) != 2 {
		t.Errorf("Expected 2 container definitions")
	}

	for _, c := range td.ContainerDefinitions {
		val := ""

		for _, e := range c.Environment {
			if *e.Name == "ECSO_ENVIRONMENT" {
				val = *e.Value
			}
		}

		if val != "test" {
			t.Errorf("ECSO_ENVIRONMENT env var not found in container %s. Want %s, got %s", *c.Name, "test", val)
		}
	}
}

func TestGetECSTaskDefinitionComposeFileError(t *testing.T) {
	env := makeTestEnvironment()
	service := makeTestService()
	service.ComposeFile = "nosuchfile"

	td, err := service.GetECSTaskDefinition(env)

	if td != nil {
		t.Error("Expected td to be nil")
	}

	if err == nil {
		t.Error("Expected error")
	}
}
func TestGetECSTaskDefinitionErrors(t *testing.T) {

	tests := []struct {
		msg       string
		configure func(*Service)
	}{
		{
			msg: "GetEnvironmentLookup failed",
			configure: func(s *Service) {
				s.environmentLookup = func(env *Environment) (lconfig.EnvironmentLookup, error) {
					return nil, fmt.Errorf("GetEnvironmentLookup failed")
				}
			},
		},
		{
			msg: "GetResourceLookup failed",
			configure: func(s *Service) {
				s.resourceLookup = func(env *Environment) (lconfig.ResourceLookup, error) {
					return nil, fmt.Errorf("GetResourceLookup failed")
				}
			},
		},
	}

	for _, test := range tests {
		service := makeTestService()
		env := makeTestEnvironment()

		test.configure(service)

		td, err := service.GetECSTaskDefinition(env)
		if err == nil {
			t.Error("Expected error")
		}

		if td != nil {
			t.Error("Expected td to be nil")
		}

		if err != nil && err.Error() != test.msg {
			t.Errorf("Want %s, got %s", test.msg, err.Error())
		}
	}
}

func TestServiceEnvironmentLookup(t *testing.T) {
	service := makeTestService()
	env := makeTestEnvironment()

	service.Environments = map[string]ServiceConfiguration{
		env.Name: ServiceConfiguration{
			Env: map[string]string{
				"FOO": "bar",
			},
		},
	}

	lookup := &ServiceEnvironmentLookup{
		Service:     service,
		Environment: env,
	}

	tests := []struct {
		input  string
		output []string
	}{
		{
			input:  "FOO",
			output: []string{"FOO=bar"},
		},
	}

	for _, test := range tests {
		got := lookup.Lookup(test.input, nil)

		if !reflect.DeepEqual(test.output, got) {
			t.Errorf("Wanted %q, got %q", test.output, got)
		}
	}
}
