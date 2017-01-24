package ecso

import (
	"fmt"
	"path"
	"path/filepath"
)

type Service struct {
	project *Project

	Name          string
	ComposeFile   string
	DesiredCount  int
	Route         string
	RoutePriority int
	Port          int
	Tags          map[string]string
	Environments  map[string]ServiceConfiguration
}

type ServiceConfiguration struct {
	Env                      map[string]string
	CloudFormationParameters map[string]string
}

func (s *Service) GetClusterName(env *Environment) string {
	return fmt.Sprintf("%s-%s", s.project.Name, env.Name)
}

func (s *Service) GetCloudFormationTemplateDir() string {
	return filepath.Join(s.project.Dir(), ".ecso", "services", s.Name)
}

func (s *Service) GetCloudFormationTemplateFile() string {
	return filepath.Join(s.GetCloudFormationTemplateDir(), "stack.yaml")
}

func (s *Service) GetCloudFormationStackName(env *Environment) string {
	return fmt.Sprintf("%s-%s-%s", s.project.Name, env.Name, s.Name)
}

func (s *Service) GetCloudFormationBucketPrefix(env *Environment) string {
	base := fmt.Sprintf("%s-%s", s.project.Name, env.Name)
	return path.Join(base, "templates", "services", s.Name)
}

func (s *Service) GetECSTaskDefinitionName(env *Environment) string {
	return fmt.Sprintf("%s-%s-%s", s.project.Name, env.Name, s.Name)
}

func (s *Service) GetECSServiceName() string {
	return fmt.Sprintf("%s-service", s.Name)
}
