package ecso

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso/resources"
	"gopkg.in/yaml.v2"
)

const (
	ecsoDotDir      = ".ecso"
	projectFilename = "project.yaml"
)

// LoadCurrentProject loads the current ecso project from the project.yaml
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

// LoadProject loads a project from the project.yaml file in the dir
// given by dir
func LoadProject(dir string) (*Project, error) {
	project := NewProject(dir, "unknown", "unknown")

	data, err := ioutil.ReadFile(project.ProjectFile())

	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, project)

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

	Name         string                  `yaml:"Name"`
	EcsoVersion  string                  `yaml:"EcsoVersion"`
	Environments map[string]*Environment `yaml:"Environments"`
	Services     map[string]*Service     `yaml:"Services"`
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

func (p *Project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type Alias Project

	aux := (*Alias)(p)

	if err := unmarshal(aux); err != nil {
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
	transform := resources.TemplateTransformation(p)

	return resources.RestoreAssetWithTransform(p.DotDir(), "project.yaml", "", transform)
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
