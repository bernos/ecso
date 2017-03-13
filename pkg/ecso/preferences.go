package ecso

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type UserPreferences struct {
	AccountDefaults map[string]AccountDefaults
}

type AccountDefaults struct {
	VPCID           string
	ALBSubnets      string
	InstanceSubnets string
	DNSZone         string
	DataDogAPIKey   string
}

func LoadCurrentUserPreferences() (*UserPreferences, error) {
	user, err := user.Current()

	if err != nil {
		return nil, err
	}

	return LoadUserPreferences(filepath.Join(user.HomeDir, ".ecso.json"))
}

func LoadUserPreferences(file string) (*UserPreferences, error) {
	preferences := &UserPreferences{}

	data, err := ioutil.ReadFile(file)

	if os.IsNotExist(err) {
		return preferences, nil
	}

	if err != nil {
		return preferences, err
	}

	err = json.Unmarshal(data, preferences)

	return preferences, err
}
