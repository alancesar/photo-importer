package file

import (
	"fmt"
	"io/ioutil"
)

const (
	PhotosDir  = "DCIM"
	VolumesDir = "/Volumes"
)

func ListVolumes() ([]string, error) {
	content, err := ioutil.ReadDir(VolumesDir)
	if err != nil {
		return nil, err
	}

	var output []string
	for _, item := range content {
		if item.IsDir() {
			output = append(output, item.Name())
		}
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("there is no device connected")
	}

	return output, nil
}
