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
	CloudFormationBucket     string
	CloudFormationParameters map[string]string
	CloudFormationTags       map[string]string
}

func (e *Environment) GetCloudFormationStackName() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}

func (e *Environment) GetCloudFormationBucketPrefix() string {
	base := fmt.Sprintf("%s-%s", e.project.Name, e.Name)
	return path.Join(base, "templates", "infrastructure")
}

func (e *Environment) GetCloudFormationTemplateDir() string {
	return filepath.Join(e.project.Dir(), ".ecso", "infrastructure", "templates")
}

func (e *Environment) GetCloudFormationTemplateFile() string {
	return filepath.Join(e.GetCloudFormationTemplateDir(), "stack.yaml")
}

func (e *Environment) GetClusterName() string {
	return fmt.Sprintf("%s-%s", e.project.Name, e.Name)
}
