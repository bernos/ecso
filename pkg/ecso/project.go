package ecso

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

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
	CloudFormationBucket     string
	CloudFormationParameters map[string]string
}

type Service struct {
	Name string
}

func (p *Project) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(p)
}

func (p *Project) AddEnvironment(name string, environment Environment) {
	if p.Environments == nil {
		p.Environments = make(map[string]Environment)
	}
	p.Environments[name] = environment
}

type UserPreferences struct {
	AccountDefaults map[string]AccountDefaults
}

type AccountDefaults struct {
	VPCID           string
	ALBSubnets      string
	InstanceSubnets string
}
