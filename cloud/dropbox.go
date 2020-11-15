package cloud

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type dropbox struct {
}

type dropboxInfo struct {
	Personal struct {
		Path string `json:"path"`
	} `json:"personal"`
}

func (dropbox) Location() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dropboxInfoLocation := filepath.Join(home, ".dropbox", "info.json")
	data, err := ioutil.ReadFile(dropboxInfoLocation)
	if err != nil {
		return "", err
	}

	info := dropboxInfo{}
	err = json.Unmarshal(data, &info)
	if err != nil {
		return "", err
	}

	return info.Personal.Path, nil
}
