package resources

import (
	"io"
	"os"
	"path/filepath"
	"text/template"
)

var (
	EnvironmentCloudFormationTemplates = &Resources{make([]Resource, 0)}
)

type Resources struct {
	rs []Resource
}

func (rs *Resources) Add(r Resource) {
	rs.rs = append(rs.rs, r)
}

func (rs *Resources) WriteTo(basePath string, data interface{}) error {
	for _, r := range rs.rs {
		if err := r.WriteTo(basePath, data); err != nil {
			return err
		}
	}
	return nil
}

type Resource interface {
	WriteTo(basePath string, data interface{}) error
}

type cloudFormationTemplate struct {
	file string
	tmpl *template.Template
}

func NewCloudFormationTemplate(file string, tmpl *template.Template) Resource {
	return &cloudFormationTemplate{
		file: file,
		tmpl: tmpl,
	}
}

func (r *cloudFormationTemplate) WriteTo(basePath string, data interface{}) error {
	return writeTemplateTo(filepath.Join(basePath, r.file), r.tmpl, data)
}

func writeTemplateTo(filename string, tmpl *template.Template, data interface{}) error {
	w, err := getWriter(filename)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func getWriter(filename string) (io.Writer, error) {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(filename)
}
