package repository

import (
	"errors"

	"gorm.io/gorm"

	"room-api/internal/model"
)

type AppVersionRepository struct {
	db *gorm.DB
}

func NewAppVersionRepository(db *gorm.DB) *AppVersionRepository {
	return &AppVersionRepository{db: db}
}

func (r *AppVersionRepository) LatestPublished() (*model.AppVersion, error) {
	var version model.AppVersion
	err := r.db.Where("is_published = ?", true).Order("version_code DESC").First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &version, nil
}
