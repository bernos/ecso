package ecso

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project"

	logrus "github.com/Sirupsen/logrus"
	lconfig "github.com/docker/libcompose/config"
)

func init() {
	// HACK The aws ecs-cli lib we use to convert the compose file to an ecs task
	// definition uses logrus directly and warns about a bunch of unsupported
	// and irrelevant compose fields. Setting logrus level here to keep it quiet
	logrus.SetLevel(logrus.ErrorLevel)
}

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

func (s *Service) Dir() string {
	return filepath.Join(s.project.Dir(), "services", s.Name)
}

func (s *Service) GetCloudFormationTemplateDir() string {
	return filepath.Join(s.project.DotDir(), "services", s.Name)
}

func (s *Service) GetCloudFormationTemplateFile() string {
	return filepath.Join(s.GetCloudFormationTemplateDir(), "stack.yaml")
}

func (s *Service) GetCloudFormationStackName(env *Environment) string {
	return fmt.Sprintf("%s-%s-%s", s.project.Name, env.Name, s.Name)
}

func (s *Service) GetDeploymentBucketPrefixForVersion(env *Environment, version string) string {
	return path.Join(s.GetDeploymentBucketPrefix(env), version)
}

func (s *Service) GetDeploymentBucketPrefix(env *Environment) string {
	return path.Join(env.GetBaseBucketPrefix(), "services", s.Name)
}

func (s *Service) GetCloudWatchLogGroup(env *Environment) string {
	return env.GetCloudWatchLogGroup()
}

func (s *Service) GetCloudWatchLogStreamPrefix(env *Environment) string {
	return fmt.Sprintf("services/%s", s.Name)
}

func (s *Service) GetECSTaskDefinitionName(env *Environment) string {
	return fmt.Sprintf("%s-%s-%s", s.project.Name, env.Name, s.Name)
}

func (s *Service) GetEnvFile(env *Environment) string {
	return filepath.Join(path.Dir(s.ComposeFile), fmt.Sprintf(".%s.env", env.Name))
}

func (s *Service) GetECSTaskDefinition(env *Environment) (*ecs.TaskDefinition, error) {
	name := s.GetECSTaskDefinitionName(env)

	envLookup, err := s.GetEnvironmentLookup(env)
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

func (s *Service) SetProject(p *Project) {
	s.project = p
}

func (s *Service) GetEnvironmentLookup(env *Environment) (*lookup.ComposableEnvLookup, error) {
	return &lookup.ComposableEnvLookup{
		Lookups: []lconfig.EnvironmentLookup{
			&lookup.EnvfileLookup{
				Path: s.GetEnvFile(env),
			},
			&EcsoServiceEnvironmentLookup{
				Service:     s,
				Environment: env,
			},
		},
	}, nil
}

type EcsoServiceEnvironmentLookup struct {
	Service     *Service
	Environment *Environment
}

func (l *EcsoServiceEnvironmentLookup) Lookup(key string, config *lconfig.ServiceConfig) []string {
	defaults := map[string]string{
		"ECSO_ENVIRONMENT":  l.Environment.Name,
		"ECSO_AWS_REGION":   l.Environment.Region,
		"ECSO_CLUSTER_NAME": l.Environment.GetClusterName(),
	}

	val, ok := l.Service.Environments[l.Environment.Name][key]
	if ok {
		return []string{fmt.Sprintf("%s=%s", key, val)}
	}

	val, ok = defaults[key]
	if ok {
		return []string{fmt.Sprintf("%s=%s", key, val)}
	}

	return []string{}
}
