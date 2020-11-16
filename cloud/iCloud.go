package cloud

import (
	"os"
	"path/filepath"
)

type iCloud struct {
}

func (iCloud) Location() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Library/Mobile Documents/com~apple~CloudDocs"), nil
}
