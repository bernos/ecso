package resources

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
)

var (
	EnvironmentCloudFormationTemplates = &Resources{make([]Resource, 0)}
	EnvironmentResources               = &Resources{make([]Resource, 0)}
)

func WriteServiceFiles(s *ecso.Service, data interface{}) error {
	fmt.Println("writing service files")
	if len(s.Route) > 0 {
		if err := WebServiceCloudFormationTemplate.WriteTo(s.GetCloudFormationTemplateDir(), data); err != nil {
			return err
		}

		if err := WebServiceDockerComposeFile.WriteTo(s.Dir(), data); err != nil {
			return err
		}
	} else {
		if err := WorkerServiceCloudFormationTemplate.WriteTo(s.GetCloudFormationTemplateDir(), data); err != nil {
			return err
		}

		if err := WorkerServiceDockerComposeFile.WriteTo(s.Dir(), data); err != nil {
			return err
		}
	}

	return nil
}

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

type textFile struct {
	file string
	tmpl *template.Template
}

func NewTextFile(file string, tmpl *template.Template) Resource {
	return &textFile{
		file: file,
		tmpl: tmpl,
	}
}

func (r *textFile) WriteTo(basePath string, data interface{}) error {
	return writeTemplateTo(filepath.Join(basePath, r.file), r.tmpl, data)
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

func NewZipFile(file string, entries map[string]*template.Template) Resource {
	return &zipFile{
		file:    file,
		entries: entries,
	}
}

type zipFile struct {
	file    string
	entries map[string]*template.Template
}

func (z *zipFile) WriteTo(basePath string, data interface{}) error {
	w, err := getWriter(filepath.Join(basePath, z.file))
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)

	for file, tmpl := range z.entries {
		f, err := zipWriter.Create(file)
		if err != nil {
			return err
		}

		if err := tmpl.Execute(f, data); err != nil {
			return err
		}
	}

	return zipWriter.Close()
}
