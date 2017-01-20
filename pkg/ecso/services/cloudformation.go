package services

import (
	"fmt"
	"regexp"
)

var (
	childTemplateRegexp = regexp.MustCompile(`(\s*)TemplateURL:\s*(\./)*(.+)`)
)

type CloudFormationService interface {
	Package(templateFile string) (string, error)
}

func findNestedTemplateFiles(templateBody string) []string {
	files := make([]string, 0)
	matches := childTemplateRegexp.FindAllStringSubmatch(templateBody, -1)

	for _, match := range matches {
		files = append(files, match[3])
	}

	return files
}

func updateNestedTemplateURLs(templateBody, region, bucket, prefix string) string {
	repl := fmt.Sprintf("${1}TemplateURL: https://s3-%s.amazonaws.com/%s/%s/$3", region, bucket, prefix)

	// repl = "${1}TemplateURL: https://bucket/foo/$3"
	return childTemplateRegexp.ReplaceAllString(templateBody, repl)
}
