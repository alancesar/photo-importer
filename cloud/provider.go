package cloud

import "fmt"

type Provider interface {
	Location() (string, error)
}

type ProviderName string

const (
	ICloudProviderName   ProviderName = "iCloud"
	OneDriveProviderName ProviderName = "OneDrive"
	DropboxProviderName  ProviderName = "Dropbox"
)

var (
	Providers = []ProviderName{OneDriveProviderName, ICloudProviderName, DropboxProviderName}
)

func NewProvider(name ProviderName) (Provider, error) {
	switch name {
	case ICloudProviderName:
		return iCloud{}, nil
	case OneDriveProviderName:
		return oneDrive{}, nil
	case DropboxProviderName:
		return dropbox{}, nil
	default:
		return nil, fmt.Errorf("invalid provider name")
	}
}
