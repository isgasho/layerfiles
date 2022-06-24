package environment

import (
	"os"
	"path/filepath"
)

func GetAndCreateDisksDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "layerfiles", "disks"), nil
}