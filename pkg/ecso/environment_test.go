package ecso

import (
	"testing"
)

func TestEnvironmentGetCloudFormationStackName(t *testing.T) {
	assertEqual("my-project-test",
		makeTestEnvironment().GetCloudFormationStackName(), t)
}

func TestEnvironmentGetCloudWatchLogGroup(t *testing.T) {
	assertEqual("my-project-test",
		makeTestEnvironment().GetCloudWatchLogGroup(), t)
}

func TestEnvironmentGetBaseBucketPrefix(t *testing.T) {
	assertEqual("my-project-test",
		makeTestEnvironment().GetBaseBucketPrefix(), t)
}

func TestEnvironmentGetDeploymentBucketPrefix(t *testing.T) {
	assertEqual("my-project-test/environment/1.0.0",
		makeTestEnvironment().GetDeploymentBucketPrefix("1.0.0"), t)
}

func TestEnvironmentResourceBucketPrefix(t *testing.T) {
	assertEqual("my-project-test/resources",
		makeTestEnvironment().GetResourceBucketPrefix(), t)
}

func TestEnvironmentGetCloudFormationTemplateDir(t *testing.T) {
	assertEqual(testDir+"/my-project/.ecso/infrastructure/templates",
		makeTestEnvironment().GetCloudFormationTemplateDir(), t)
}

func TestEnvironmentGetResourceDir(t *testing.T) {
	assertEqual(testDir+"/my-project/.ecso/infrastructure/templates/resources",
		makeTestEnvironment().GetResourceDir(), t)
}

func TestEnvironmentGetCloudFormationTemplateFile(t *testing.T) {
	assertEqual(testDir+"/my-project/.ecso/infrastructure/templates/stack.yaml",
		makeTestEnvironment().GetCloudFormationTemplateFile(), t)
}

func TestEnvironmentGetClusterName(t *testing.T) {
	assertEqual("my-project-test",
		makeTestEnvironment().GetClusterName(), t)
}

func TestEnvironmentSetProject(t *testing.T) {
	project := &Project{}
	env := &Environment{}
	env.SetProject(project)

	assertEqual(project, env.project, t)
}
