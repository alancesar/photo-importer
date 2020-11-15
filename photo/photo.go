package photo

import "gorm.io/gorm"

type Photo struct {
	gorm.Model
	Filename string
	Checksum string
}
