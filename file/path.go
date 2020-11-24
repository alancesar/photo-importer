package file

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
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

func FindImagesDirectory(rootPath string) (string, error) {
	content, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return "", err
	}

	compile, err := regexp.Compile("\\d{3}.*")
	if err != nil {
		return "", err
	}

	for _, item := range content {
		if item.IsDir() {
			if compile.MatchString(item.Name()) {
				return filepath.Join(rootPath, item.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("DCIM directory not valid")
}

func GetDeviceUUID(device string) (string, error) {
	cmd := fmt.Sprintf("diskutil info %s | grep UUID", device)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.Split(string(out), ":")[1:][0]), nil
}
