package model

import "time"

type Room struct {
	ID         uint64    `gorm:"primaryKey;column:id"`
	Name       string    `gorm:"column:name"`
	OwnerID    uint64    `gorm:"column:owner_id"`
	MaxMembers uint8     `gorm:"column:max_members"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Room) TableName() string {
	return "rooms"
}
