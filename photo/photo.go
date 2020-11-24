package photo

import (
	"github.com/alancesar/photo-importer/cloud"
	"gorm.io/gorm"
)

type Photo struct {
	gorm.Model
	Filename   string
	Checksum   string
	DeviceUUID string
	Provider   cloud.ProviderName
}

func (p Photo) Exists() bool {
	return p.ID != 0
}
