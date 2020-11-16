package cloud

import (
	"os"
	"path/filepath"
)

type oneDrive struct {
}

func (oneDrive) Location() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "OneDrive"), nil
}
