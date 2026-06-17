package model

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	Nickname     string    `gorm:"column:nickname"`
	PasswordHash string    `gorm:"column:password_hash"`
	AvatarURL    string    `gorm:"column:avatar_url"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
