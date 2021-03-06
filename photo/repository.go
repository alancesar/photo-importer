package photo

import (
	"github.com/alancesar/photo-importer/cloud"
	"gorm.io/gorm"
)

type Repository interface {
	Get(filename, deviceUUID string, provider cloud.ProviderName) (Photo, error)
	Save(photo *Photo) error
	Delete(photo Photo) error
}

type sqlite struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) (Repository, error) {
	if err := db.AutoMigrate(&Photo{}); err != nil {
		return nil, err
	}

	return sqlite{db: db}, nil
}

func (s sqlite) Get(filename, deviceUUID string, provider cloud.ProviderName) (Photo, error) {
	p := Photo{}
	query := s.db.Where("filename = ? AND device_uuid = ? AND provider = ?", filename, deviceUUID, provider).
		Last(&p)

	if query.Error != nil && query.Error != gorm.ErrRecordNotFound {
		return Photo{}, query.Error
	}

	return p, nil
}

func (s sqlite) Save(p *Photo) error {
	query := s.db.Create(p)
	return query.Error
}

func (s sqlite) Delete(p Photo) error {
	query := s.db.Delete(&p, "filename = ? AND checksum = ? AND provider = ?",
		p.Filename, p.Checksum, p.Provider)

	return query.Error
}
