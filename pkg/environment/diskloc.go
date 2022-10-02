package environment

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func GetAndCreateDisksDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	res := filepath.Join(homeDir, ".layerfiles", "disks")
	err = os.MkdirAll(res, 0750)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrapf(err, "could not create disk directory at %v", res)
	}
	return res, nil
}

func GetAndCreateBinDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	res := filepath.Join(homeDir, ".layerfiles", "bin")
	err = os.MkdirAll(res, 0700)
	if err != nil && !os.IsExist(err) {
		return "", errors.Wrapf(err, "could not create bin directory at %v", res)
	}
	return res, nil
}
