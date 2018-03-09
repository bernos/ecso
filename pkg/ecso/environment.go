package ecso

import (
	"fmt"
	"path"
	"path/filepath"
)

var (
	// EnvironmentBaseDir is the path relative to the .ecso dir that contains
	// all environment files
	EnvironmentBaseDir = filepath.Join(ecsoDotDir, "environment")

	// EnvironmentCloudFormationDir is the path relative to the .ecso dir that the
	// environment cloudformation templates are stored
	EnvironmentCloudFormationDir = filepath.Join(EnvironmentBaseDir, "cloudformation")

	// EnvironmentResourceDir is the path relative to the .ecso dir that the
	// environment resource files are stored
	EnvironmentResourceDir = filepath.Join(EnvironmentCloudFormationDir, "resources")

	// EnvironmentCloudFormationTemplateFile is the filename of the root
	// cloudformation template for an environment
	EnvironmentCloudFormationTemplateFile = filepath.Join(EnvironmentCloudFormationDir, "stack.yaml")

	DefaultEnvironmentName = "dev"

	DefaultRegion = "ap-southeast-2"
)

type Environment struct {
	project *Project

	Name                     string            `yaml:"Name"`
	Region                   string            `yaml:"Region"`
	CloudFormationParameters map[string]string `yaml:"CloudFormationParameters"`
	CloudFormationTags       map[string]string `yaml:"CloudFormationTags"`
}

func (e *Environment) GetCloudFormationStackName() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}

func (e *Environment) GetCloudWatchLogGroup() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}

func (e *Environment) GetBaseBucketPrefix() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}

func (e *Environment) GetDeploymentBucketPrefix(version string) string {
	return path.Join(e.GetBaseBucketPrefix(), "environment", version)
}

func (e *Environment) GetResourceBucketPrefix() string {
	return path.Join(e.GetBaseBucketPrefix(), "resources")
}

func (e *Environment) GetCloudFormationTemplateDir() string {
	return filepath.Join(e.project.Dir(), EnvironmentCloudFormationDir)
}

func (e *Environment) GetResourceDir() string {
	return filepath.Join(e.project.Dir(), EnvironmentResourceDir)
}

func (e *Environment) GetCloudFormationTemplateFile() string {
	return filepath.Join(e.project.Dir(), EnvironmentCloudFormationTemplateFile)
}

func (e *Environment) GetClusterName() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}

func (e *Environment) SetProject(p *Project) {
	e.project = p
}
