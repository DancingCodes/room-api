package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"room-api/internal/model"
)

type AppVersionRepository struct {
	db *gorm.DB
}

func (r *AppVersionRepository) List() ([]model.AppVersion, error) {
	var versions []model.AppVersion
	if err := r.db.Order("version_code DESC").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *AppVersionRepository) FindByID(id uint64) (*model.AppVersion, error) {
	var version model.AppVersion
	if err := r.db.First(&version, id).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *AppVersionRepository) Create(versionCode uint64, apkURL, releaseNotes string) (*model.AppVersion, error) {
	now := time.Now()
	version := model.AppVersion{
		VersionCode:  versionCode,
		APKURL:       apkURL,
		ReleaseNotes: releaseNotes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := r.db.Create(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *AppVersionRepository) UpdateDraft(id, versionCode uint64, apkURL, releaseNotes string) (*model.AppVersion, error) {
	var version model.AppVersion
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&version, id).Error; err != nil {
			return err
		}
		if version.IsPublished {
			return errors.New("已发布版本不能编辑")
		}
		if err := tx.Model(&version).Updates(map[string]any{
			"version_code":  versionCode,
			"apk_url":       apkURL,
			"release_notes": releaseNotes,
			"updated_at":    time.Now(),
		}).Error; err != nil {
			return err
		}
		version.VersionCode = versionCode
		version.APKURL = apkURL
		version.ReleaseNotes = releaseNotes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *AppVersionRepository) Publish(id uint64) (*model.AppVersion, error) {
	var version model.AppVersion
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&version, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.AppVersion{}).Where("is_published = ?", true).Update("is_published", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&version).Updates(map[string]any{"is_published": true, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		version.IsPublished = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *AppVersionRepository) Unpublish(id uint64) error {
	result := r.db.Model(&model.AppVersion{}).Where("id = ?", id).Updates(map[string]any{"is_published": false, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
