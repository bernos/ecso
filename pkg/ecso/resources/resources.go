package resources

//go:generate go-bindata -ignore=.*node_modules -pkg $GOPACKAGE -o resources-generated.go ./...

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bernos/ecso/pkg/ecso"
)

var (
	InstanceDrainerLambdaVersion = "1.0.0"

	ServiceDiscoveryLambdaVersion = "1.0.0"

	DNSCleanerLambdaVersion = "1.0.0"

	EnvironmentFiles = environmentFiles()
)

func environmentFiles() []Resource {
	files := environmentCfnTemplates()
	files = append(files, environmentLambdas()...)

	return files
}

func environmentLambdas() []Resource {
	zip := func(name, version string) string {
		return fmt.Sprintf("%s/lambda/%s-%s.zip", ecso.EnvironmentResourceDir, name, version)
	}

	src := func(lambda, path string) string {
		return filepath.Join("environment", "lambda", lambda, path)
	}

	return []Resource{
		NewZipFile(zip("instance-drainer", InstanceDrainerLambdaVersion),
			MustParseTemplateAsset("index.py", src("instance-drainer", "index.py"))),

		NewZipFile(zip("service-discovery", ServiceDiscoveryLambdaVersion),
			MustParseTemplateAsset("index.js", src("service-discovery", "index.js"))),

		NewZipFile(zip("dns-cleaner", DNSCleanerLambdaVersion),
			MustParseTemplateAsset("index.js", src("dns-cleaner", "index.js"))),
	}
}

func environmentCfnTemplates() []Resource {
	cfnTemplates := []string{
		"alarms.yaml",
		"dd-agent.yaml",
		"dns-cleaner.yaml",
		"ecs-cluster.yaml",
		"instance-drainer.yaml",
		"load-balancers.yaml",
		"logging.yaml",
		"security-groups.yaml",
		"service-discovery.yaml",
		"sns.yaml",
		"stack.yaml",
	}

	src := func(path string) string {
		return filepath.Join("environment", "cloudformation", path)
	}

	dst := func(path string) string {
		return filepath.Join(ecso.EnvironmentCloudFormationDir, path)
	}

	files := make([]Resource, 0)

	for _, cfnTemplate := range cfnTemplates {
		files = append(files, NewTextFile(MustParseTemplateAsset(dst(cfnTemplate), src(cfnTemplate))))
	}

	return files
}

type Resource interface {
	Filename() string
	WriteTo(w io.Writer, data interface{}) error
}

func MustParseTemplateAsset(name, assetPath string) *template.Template {
	a := MustAsset(assetPath)
	t := template.Must(template.New(name).Parse(string(a)))

	return t
}

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
