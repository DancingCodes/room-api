package model

import "time"

type RoomMember struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	RoomID    uint64    `gorm:"column:room_id"`
	UserID    uint64    `gorm:"column:user_id"`
	IsOwner   bool      `gorm:"column:is_owner"`
	MicStatus string    `gorm:"column:mic_status"`
	JoinedAt  time.Time `gorm:"column:joined_at"`
}

func (RoomMember) TableName() string {
	return "room_members"
}
