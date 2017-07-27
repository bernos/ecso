package resources

import (
	"io"
	"os"
	"path/filepath"
	"text/template"
)

type textFile struct {
	*template.Template
}

func NewTextFile(tmpl *template.Template) Resource {
	return &textFile{tmpl}
}

func (r *textFile) Filename() string {
	return r.Template.Name()
}

func (r *textFile) WriteTo(w io.Writer, data interface{}) error {
	return r.Template.Execute(w, data)
}

type ResourceWriter interface {
	WriteResource(r Resource, data interface{}) error
	WriteResources(data interface{}, rs ...Resource) error
}

type FileSystemResourceWriter struct {
	dir string
}

func NewFileSystemResourceWriter(dir string) ResourceWriter {
	return &FileSystemResourceWriter{dir}
}

func (w *FileSystemResourceWriter) WriteResource(r Resource, data interface{}) error {
	filename := filepath.Join(w.dir, r.Filename())

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	return r.WriteTo(f, data)
}

func (w *FileSystemResourceWriter) WriteResources(data interface{}, rs ...Resource) error {
	for _, r := range rs {
		if err := w.WriteResource(r, data); err != nil {
			return err
		}
	}
	return nil
}
