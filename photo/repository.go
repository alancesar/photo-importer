package photo

import "gorm.io/gorm"

type Repository interface {
	Get(filename, checksum string) (Photo, error)
	Save(photo *Photo) error
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

func (s sqlite) Get(filename, checksum string) (Photo, error) {
	p := Photo{}
	query := s.db.Where("filename = ? AND checksum = ?", filename, checksum).
		First(&p)

	if query.Error != nil && query.Error != gorm.ErrRecordNotFound {
		return Photo{}, query.Error
	}

	return p, nil
}

func (s sqlite) Save(p *Photo) error {
	query := s.db.Create(p)
	return query.Error
}