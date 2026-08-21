package model

import "time"

type AppVersion struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	VersionCode  uint64    `gorm:"column:version_code"`
	APKURL       string    `gorm:"column:apk_url"`
	ReleaseNotes string    `gorm:"column:release_notes"`
	IsPublished  bool      `gorm:"column:is_published"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (AppVersion) TableName() string {
	return "app_versions"
}
