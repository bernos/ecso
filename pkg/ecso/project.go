package ecso

import (
	"encoding/json"
	"io"
)

func NewProject(name string) *Project {
	return &Project{
		Name: name,
	}
}

type Project struct {
	Name         string
	Environments []Environment
	Services     []Service
}

type Environment struct {
	Name string
}

type Service struct {
	Name string
}

func (p *Project) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(p)
}
