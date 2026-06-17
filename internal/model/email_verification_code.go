package model

import "time"

type EmailVerificationCode struct {
	ID        uint64     `gorm:"primaryKey;column:id"`
	UserID    *uint64    `gorm:"column:user_id"`
	Email     string     `gorm:"column:email"`
	Purpose   string     `gorm:"column:purpose"`
	Code      string     `gorm:"column:code"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (EmailVerificationCode) TableName() string {
	return "email_verification_codes"
}
