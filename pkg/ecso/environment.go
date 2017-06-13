package ecso

import (
	"fmt"
	"path"
	"path/filepath"
)

type Environment struct {
	project *Project

	Name                     string
	Region                   string
	CloudFormationParameters map[string]string
	CloudFormationTags       map[string]string
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

func (e *Environment) GetCloudFormationBucketPrefix() string {
	return path.Join(e.GetBaseBucketPrefix(), "templates", "infrastructure")
}

func (e *Environment) GetResourceBucketPrefix() string {
	return path.Join(e.GetBaseBucketPrefix(), "resources")
}

func (e *Environment) GetCloudFormationTemplateDir() string {
	return filepath.Join(e.project.Dir(), ".ecso", "infrastructure", "templates")
}

func (e *Environment) GetResourceDir() string {
	return filepath.Join(e.GetCloudFormationTemplateDir(), "resources")
}

func (e *Environment) GetCloudFormationTemplateFile() string {
	return filepath.Join(e.GetCloudFormationTemplateDir(), "stack.yaml")
}

func (e *Environment) GetClusterName() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}
