package file

import (
	"github.com/alancesar/tidy-photo/exif"
	"os"
	"path/filepath"
	"strings"
)

const (
	pattern         = "2006/2006-01-02"
	separator       = "/"
	rootDestination = "OneDrive/Photos"
)

func BuildFilename(filename string, parser *exif.Parser) (string, error) {
	time, err := parser.GetDateTime()
	if err != nil {
		return "", err
	}

	date := time.Format(pattern)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	destination := filepath.Join(strings.Split(date, separator)...)
	destination = filepath.Join(homeDir, rootDestination, destination, filename)
	destination = filepath.Clean(destination)
	return destination, nil
}
