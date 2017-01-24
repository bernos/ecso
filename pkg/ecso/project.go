package ecso

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

func GetCurrentProjectDir() (string, error) {
	// For now this is just pwd, but later might want to walk up
	// the dir tree, so ecso can run from sub folders in a project
	return os.Getwd()
}

func LoadProject(dir string) (*Project, error) {
	project := NewProject(dir, "unknown")

	data, err := ioutil.ReadFile(project.ProjectFile())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, project)

	return project, err
}

func NewProject(dir, name string) *Project {
	return &Project{
		dir:          dir,
		Name:         name,
		Environments: make(map[string]*Environment),
		Services:     make(map[string]*Service),
	}
}

type Project struct {
	dir string

	Name         string
	Environments map[string]*Environment
	Services     map[string]*Service
}

func (p *Project) Dir() string {
	return p.dir
}

func (p *Project) ProjectFile() string {
	return filepath.Join(p.Dir(), ".ecso", "project.json")
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
		// p.Environments[i] = env
	}

	for _, svc := range p.Services {
		svc.project = p
		// p.Services[i] = svc
	}

	return nil
}

func (p *Project) Save() error {
	w, err := os.Create(p.ProjectFile())

	if err != nil {
		return err
	}

	return p.Write(w)
}

func (p *Project) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(p)
}

func (p *Project) AddEnvironment(environment *Environment) {
	if p.Environments == nil {
		p.Environments = make(map[string]*Environment)
	}
	p.Environments[environment.Name] = environment
}

func (p *Project) AddService(service *Service) {
	if p.Services == nil {
		p.Services = make(map[string]*Service)
	}
	p.Services[service.Name] = service
}

type UserPreferences struct {
	AccountDefaults map[string]AccountDefaults
}

type AccountDefaults struct {
	VPCID           string
	ALBSubnets      string
	InstanceSubnets string
}
