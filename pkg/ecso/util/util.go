package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/bernos/ecso/pkg/ecso"
)

func LoadUserPreferences() (ecso.UserPreferences, error) {
	preferences := ecso.UserPreferences{}
	user, err := user.Current()

	if err != nil {
		return preferences, err
	}

	data, err := ioutil.ReadFile(filepath.Join(user.HomeDir, ".ecso.json"))

	if os.IsNotExist(err) {
		return preferences, nil
	}

	if err != nil {
		return preferences, err
	}

	err = json.Unmarshal(data, &preferences)

	return preferences, err
}

func LoadCurrentProject() (*ecso.Project, error) {
	file, err := GetCurrentProjectFile()

	if err != nil {
		return nil, err
	}

	project, err := ecso.LoadProject(file)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("No ecso project file was found at %s", file)
	}

	return project, err
}

func SaveCurrentProject(project *ecso.Project) error {
	file, err := GetCurrentProjectFile()

	if err != nil {
		return err
	}

	w, err := os.Create(file)

	if err != nil {
		return err
	}

	return project.Save(w)
}

func GetCurrentProjectFile() (string, error) {
	wd, err := os.Getwd()

	return filepath.Join(wd, ".ecso", "project.json"), err
}
