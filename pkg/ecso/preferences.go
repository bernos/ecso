package ecso

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

func LoadUserPreferences() (UserPreferences, error) {
	preferences := UserPreferences{}
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
