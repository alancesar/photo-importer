package cloud

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type dropbox struct {
	infoDataReader func() ([]byte, error)
}

type dropboxInfo struct {
	Personal struct {
		Path string `json:"path"`
	} `json:"personal"`
}

func (dropbox) defaultInfoDataReader() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dropboxInfoLocation := filepath.Join(home, ".dropbox", "info.json")
	return ioutil.ReadFile(dropboxInfoLocation)
}

func (d dropbox) Location() (string, error) {
	if d.infoDataReader == nil {
		d.infoDataReader = d.defaultInfoDataReader
	}

	data, err := d.infoDataReader()
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
