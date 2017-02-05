package api

import "github.com/bernos/ecso/pkg/ecso"

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
	EnvironmentUp(p *ecso.Project, env *ecso.Environment, dryRun bool) error
	EnvironmentDown(p *ecso.Project, env *ecso.Environment) error
	// EnvironmentRemove()

	ServiceUp(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	ServiceDown(p *ecso.Project, env *ecso.Environment, s *ecso.Service) error
	// ServiceDown()

	// ListTasks()
	// ListServices()

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
