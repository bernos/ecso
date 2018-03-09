package resources

//go:generate go-bindata -ignore=.*node_modules -pkg $GOPACKAGE -o resources-generated.go ./...

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func TemplateTransformation(data interface{}) func(content []byte, name string) ([]byte, error) {
	return func(content []byte, name string) ([]byte, error) {
		t, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer

		if err := t.Execute(&buf, data); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}

// RestoreAsset restores an asset under the given directory
func RestoreAssetWithTransform(dir, name, trimPrefix string, transform func([]byte, string) ([]byte, error)) error {
	trimmed := strings.TrimPrefix(name, trimPrefix)

	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(trimmed)), os.FileMode(0755))
	if err != nil {
		return err
	}
	transformed, err := transform(data, name)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, trimmed), transformed, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, trimmed), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssetsWithTransform(dir, name, trimPrefix string, transform func([]byte, string) ([]byte, error)) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAssetWithTransform(dir, name, trimPrefix, transform)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssetsWithTransform(dir, filepath.Join(name, child), trimPrefix, transform)
		if err != nil {
			return err
		}
	}
	return nil
}

func RestoreAssetDirWithTransform(dir, name string, transform func([]byte, string) ([]byte, error)) error {
	return RestoreAssetsWithTransform(dir, name, name, transform)
}
