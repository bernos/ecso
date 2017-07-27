package resources

import (
	"html/template"
	"testing"
)

func TestTemplateParseName(t *testing.T) {
	tests := []string{
		"test.yaml",
		"a/b/c.yaml",
	}

	for _, test := range tests {
		tmpl := template.Must(template.New(test).Parse(environmentALBTemplate))

		if tmpl.Name() != test {
			t.Errorf("Want '%s', got '%s'.", test, tmpl.Name())
		}
	}

}
