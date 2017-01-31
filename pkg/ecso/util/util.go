package util

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
)

func DirExists(dir string) (bool, error) {
	_, err := os.Stat(dir)

	switch {
	case os.IsNotExist(err):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func WriteFileFromTemplate(filename string, tmpl *template.Template, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	w, err := os.Create(filename)

	if err != nil {
		return err
	}

	return tmpl.Execute(w, data)
}

func AnyError(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}

func ClusterConsoleURL(cluster, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/home?region=%s#/clusters/%s/services", region, region, cluster)
}

func ServiceConsoleURL(service, cluster, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/home?region=%s#/clusters/%s/services/%s/tasks", region, region, cluster, service)
}

func CloudFormationConsoleURL(stackID, region string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stack/detail?stackId=%s", region, region, url.QueryEscape(stackID))
}
