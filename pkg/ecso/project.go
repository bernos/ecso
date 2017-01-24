package ecso

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func LoadCurrentProject() (*Project, error) {
	file, err := GetCurrentProjectFile()

	if err != nil {
		return nil, err
	}

	project, err := LoadProject(file)

	if os.IsNotExist(err) {
		return project, nil
	}

	return project, err
}

func SaveCurrentProject(project *Project) error {
	file, err := GetCurrentProjectFile()

	if err != nil {
		return err
	}

	w, err := os.Create(file)

	if err != nil {
		return err
	}

	return project.Save(w)
}

func GetCurrentProjectFile() (string, error) {
	wd, err := GetCurrentProjectDir()

	return filepath.Join(wd, ".ecso", "project.json"), err
}

func GetCurrentProjectDir() (string, error) {
	// For now this is just pwd, but later might want to walk up
	// the dir tree, so ecso can run from sub folders in a project
	return os.Getwd()
}

func LoadProject(path string) (*Project, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var project Project

	err = json.Unmarshal(data, &project)

	return &project, err
}

func NewProject(name string) *Project {
	return &Project{
		Name: name,
	}
}

type Project struct {
	Name         string
	Environments map[string]Environment
	Services     map[string]Service
}

type Environment struct {
	Name                     string
	Region                   string
	CloudFormationBucket     string
	CloudFormationParameters map[string]string
	CloudFormationTags       map[string]string
}

func (p *Project) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(p)
}

func (p *Project) AddEnvironment(environment Environment) {
	if p.Environments == nil {
		p.Environments = make(map[string]Environment)
	}
	p.Environments[environment.Name] = environment
}

func (p *Project) AddService(service Service) {
	if p.Services == nil {
		p.Services = make(map[string]Service)
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
