package ecso

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ecsoDotDir      = ".ecso"
	projectFilename = "project.json"
)

// LoadCurrentProject loads the current ecso project from the project.json
// file located at the dir given by GetCurrentProjectDir()
func LoadCurrentProject() (*Project, error) {
	dir, err := GetCurrentProjectDir()

	if err != nil {
		return nil, err
	}

	project, err := LoadProject(dir)

	if os.IsNotExist(err) {
		return project, nil
	}

	return project, err
}

// GetCurrentProjectDir locates the root directory of the current ecso
// project.
// For now this is just pwd, but later might want to walk up
// the dir tree, so ecso can run from sub folders in a project
func GetCurrentProjectDir() (string, error) {
	return os.Getwd()
}

// LoadProject loads a project from the project.json file in the dir
// given by dir
func LoadProject(dir string) (*Project, error) {
	project := NewProject(dir, "unknown", "unknown")

	data, err := ioutil.ReadFile(project.ProjectFile())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, project)

	return project, err
}

// NewProject creates a new project
func NewProject(dir, name, version string) *Project {
	return &Project{
		dir:          dir,
		Name:         name,
		EcsoVersion:  version,
		Environments: make(map[string]*Environment),
		Services:     make(map[string]*Service),
	}
}

// Project is a container for all environment and service configurations
// Projects are saved to a project.json file located in the `.ecso` dir of
// the project directory
type Project struct {
	dir string

	Name         string
	EcsoVersion  string
	Environments map[string]*Environment
	Services     map[string]*Service
}

func (p *Project) Dir() string {
	return p.dir
}

func (p *Project) DotDir() string {
	return filepath.Join(p.Dir(), ecsoDotDir)
}

func (p *Project) HasEnvironment(name string) bool {
	return p.Environments[name] != nil
}

func (p *Project) HasService(name string) bool {
	return p.Services[name] != nil
}

func (p *Project) ProjectFile() string {
	return filepath.Join(p.DotDir(), projectFilename)
}

func (p *Project) UnmarshalJSON(b []byte) error {
	type Alias Project

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}

	for _, env := range p.Environments {
		env.project = p
	}

	for _, svc := range p.Services {
		svc.project = p
	}

	return nil
}

func (p *Project) Save() error {
	w, err := os.Create(p.ProjectFile())
	if err != nil {
		return err
	}

	_, err = p.WriteTo(w)

	return err
}

func (p *Project) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(b)

	return int64(n), err
}

func (p *Project) AddEnvironment(environment *Environment) {
	if p.Environments == nil {
		p.Environments = make(map[string]*Environment)
	}
	p.Environments[environment.Name] = environment
	environment.project = p
}

func (p *Project) AddService(service *Service) {
	if p.Services == nil {
		p.Services = make(map[string]*Service)
	}
	p.Services[service.Name] = service
	service.project = p
}
