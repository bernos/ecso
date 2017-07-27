package resources

import (
	"archive/zip"
	"io"
	"text/template"
)

type zipFile struct {
	file    string
	entries []*template.Template
}

func NewZipFile(file string, entries ...*template.Template) Resource {
	return &zipFile{
		file:    file,
		entries: entries,
	}
}

func (z *zipFile) Filename() string {
	return z.file
}

func (z *zipFile) WriteTo(w io.Writer, data interface{}) error {
	zipWriter := zip.NewWriter(w)

	for _, tmpl := range z.entries {
		f, err := zipWriter.Create(tmpl.Name())
		if err != nil {
			return err
		}

		if err := tmpl.Execute(f, data); err != nil {
			return err
		}
	}

	return zipWriter.Close()
}
