package api

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
)

type API interface {
	DescribeEnvironment(env *ecso.Environment) (*EnvironmentDescription, error)
	DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error)

	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment) error

	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceLogs(p *ecso.Project, env *ecso.Environment, s *ecso.Service) ([]*cloudwatchlogs.FilteredLogEvent, error)

	GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error)
}

// New creates a new API
func New(cfg *ecso.Config) API {
	return &api{cfg}
}

type api struct {
	cfg *ecso.Config
}
