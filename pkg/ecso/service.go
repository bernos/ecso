package ecso

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
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
	return fmt.Sprintf("%s", s.Name)
}

func (s *Service) GetECSTaskDefinition(env *Environment) (*ecs.TaskDefinition, error) {

	name := s.GetECSTaskDefinitionName(env)

	envLookup, err := utils.GetDefaultEnvironmentLookup()

	if err != nil {
		return nil, err
	}

	resourceLookup, err := utils.GetDefaultResourceLookup()

	if err != nil {
		return nil, err
	}

	context := &project.Context{
		ComposeFiles:      []string{s.ComposeFile},
		ProjectName:       name,
		EnvironmentLookup: envLookup,
		ResourceLookup:    resourceLookup,
	}

	p := project.NewProject(context, nil, nil)

	if err := p.Parse(); err != nil {
		return nil, err
	}

	return utils.ConvertToTaskDefinition(name, context, p.ServiceConfigs)
}
