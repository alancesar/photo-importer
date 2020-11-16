package file

import (
	"fmt"
	"path/filepath"
)

type CheckFunc func(path string) (bool, error)
type SumFunc func(path string) (string, error)

type Handler struct {
	exists CheckFunc
	sum    SumFunc
}

func NewHandler(exists CheckFunc, sum SumFunc) Handler {
	return Handler{
		exists: exists,
		sum:    sum,
	}
}

func (h Handler) IsDuplicated(output, checksum string) (duplicated bool, newPath string, err error) {
	index := 0
	newPath = output

	for {
		index++
		exists, err := h.exists(newPath)
		if err != nil {
			return false, "", err
		} else if !exists {
			return false, newPath, nil
		}

		destChecksum, err := h.sum(newPath)
		if err != nil {
			return false, "", err
		}

		if checksum == destChecksum {
			return true, "", nil
		}

		newPath = output[:len(output)-len(filepath.Ext(output))]
		newPath = fmt.Sprintf("%s_%d%s", newPath, index, filepath.Ext(output))
	}
}
