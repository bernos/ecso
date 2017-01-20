package util

import "os"

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
