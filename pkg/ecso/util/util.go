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
	wd, err := GetCurrentProjectDir()

	return filepath.Join(wd, ".ecso", "project.json"), err
}

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

func GetCurrentProjectDir() (string, error) {
	// For now this is just pwd, but later might want to walk up
	// the dir tree, so ecso can run from sub folders in a project
	return os.Getwd()
}
