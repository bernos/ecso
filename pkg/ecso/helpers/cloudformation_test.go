package helpers

import (
	"io/ioutil"
	"testing"
)

func TestFindNestedTemplateFiles(t *testing.T) {
	wants := []string{
		"infrastructure/security-groups.yaml",
		"infrastructure/load-balancers.yaml",
		"infrastructure/ecs-cluster.yaml",
	}

	gots := findNestedTemplateFiles(MustReadFile(t, "./testdata/root_template.yaml"))

	if len(wants) != len(gots) {
		t.Errorf("Want %v, got %v", wants, gots)
	} else {
		for i, want := range wants {
			if want != gots[i] {
				t.Errorf("Want %s, got %s", want, gots[i])
			}
		}
	}
}

func TestUpdateNestedTemplateURLs(t *testing.T) {
	want := MustReadFile(t, "./testdata/packaged_root_template.yaml")
	got := updateNestedTemplateURLs(MustReadFile(t, "./testdata/root_template.yaml"), "ap-southeast-2", "bucketname", "my/bucket/prefix")

	if want != got {
		t.Errorf("Want %s, got %s", want, got)
	}
}

func MustReadFile(t *testing.T, filename string) string {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		t.Fatalf("Failed to ReadFile(%s) : %s", filename, err.Error())
	}

	return string(data)
}
