package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"room-api/internal/model"
)

type EmailCodeRepository struct {
	db *gorm.DB
}

func NewEmailCodeRepository(db *gorm.DB) *EmailCodeRepository {
	return &EmailCodeRepository{db: db}
}

func (r *EmailCodeRepository) CountSince(email, purpose string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.EmailVerificationCode{}).
		Where("email = ? AND purpose = ? AND created_at >= ?", email, purpose, since).
		Count(&count).Error
	return count, err
}

func (r *EmailCodeRepository) InvalidateActive(email, purpose string, now time.Time) error {
	return r.db.Model(&model.EmailVerificationCode{}).
		Where("email = ? AND purpose = ? AND used_at IS NULL", email, purpose).
		Update("used_at", now).Error
}

func (r *EmailCodeRepository) Create(code *model.EmailVerificationCode) error {
	return r.db.Create(code).Error
}

func (r *EmailCodeRepository) LatestActive(email, purpose string, now time.Time) (*model.EmailVerificationCode, error) {
	var code model.EmailVerificationCode
	err := r.db.
		Where("email = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?", email, purpose, now).
		Order("created_at DESC").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *EmailCodeRepository) MarkUsed(id uint64, usedAt time.Time) error {
	return r.db.Model(&model.EmailVerificationCode{}).Where("id = ?", id).Update("used_at", usedAt).Error
}
