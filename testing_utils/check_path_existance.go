package testing_utils

import (
	"fmt"

	"github.com/francoishill/afero"
)

func CheckFileSystemPathExistance(fs afero.Fs, relPath string, mustExist bool) error {
	exists, err := afero.Exists(fs, relPath)
	if err != nil {
		return fmt.Errorf("Cannot check if path '%s' exists, error: %s", relPath, err.Error())
	}

	if exists && !mustExist {
		return fmt.Errorf("Path '%s' existed and it should not have", relPath)
	} else if !exists && mustExist {
		return fmt.Errorf("Path '%s' did not exist but it should have", relPath)
	}
	return nil
}
