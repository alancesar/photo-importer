package prompt

import (
	"github.com/manifoldco/promptui"
)

func SelectDevices(volumes []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select the source device",
		Items: volumes,
	}

	_, device, err := prompt.Run()
	return device, err
}
