package ecso

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/ecs"
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

func (s *Service) GetECSTaskDefinition(env *Environment) (*ecs.TaskDefinition, error) {

	return &ecs.TaskDefinition{}, nil
	// envLookup, err := util.GetDefaultEnvironmentLookup()

	// if err != nil {
	// 	return nil, err
	// }

	// resourceLookup, err := util.GetDefaultResourceLookup()

	// if err != nil {
	// 	return nil, err
	// }

	// context := &ctx.Context{
	// 	Context: project.Context{
	// 		ComposeFiles:      []string{s.ComposeFile},
	// 		ProjectName:       s.GetECSTaskDefinitionName(env),
	// 		EnvironmentLookup: envLookup,
	// 		ResourceLookup:    resourceLookup,
	// 	},
	// }

	// p, err := docker.NewProject(context, nil)

	// if err != nil {
	// 	return nil, err
	// }

	// serviceConfigs := p.(*project.Project).ServiceConfigs

	// return util.ConvertToTaskDefinition(s.GetECSTaskDefinitionName(env), &context.Context, serviceConfigs)
}
