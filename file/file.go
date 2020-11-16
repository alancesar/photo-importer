package file

import (
	"path/filepath"
	"strings"
	"time"
)

const (
	pattern   = "2006/2006-01-02"
	separator = "/"
)

func DateToPath(time time.Time) string {
	date := time.Format(pattern)
	return filepath.Join(strings.Split(date, separator)...)
}
