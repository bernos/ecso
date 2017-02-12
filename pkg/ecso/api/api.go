package api

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
)

// env

// environment add
// environment up
// environment rm

// service add
// service up
// service down
// service ls
// service ps
// service logs

type API interface {
	DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error)
	DescribeService(env *ecso.Environment, service *ecso.Service) (map[string]string, error)

	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment) error

	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error)

	GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error)
	// ListTasks()

	// GetLogs()

	// LoadProject(dir string) (*ecso.Project, error)
	// SaveProject(p *ecso.Project) error
}

// New creates a new API
func New(cfg *ecso.Config) API {
	return &api{cfg}
}

type api struct {
	cfg *ecso.Config
}
