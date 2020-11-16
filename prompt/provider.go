package prompt

import (
	"github.com/alancesar/photo-importer/cloud"
	"github.com/manifoldco/promptui"
)

func ProviderName(availableProviders []cloud.ProviderName) (cloud.ProviderName, error) {
	prompt := promptui.Select{
		Label: "Select the cloud provider",
		Items: availableProviders,
	}

	_, provider, err := prompt.Run()
	return cloud.ProviderName(provider), err
}
