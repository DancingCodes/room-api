package model

import "time"

type Message struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	RoomID    uint64    `gorm:"column:room_id"`
	SenderID  uint64    `gorm:"column:sender_id"`
	Type      string    `gorm:"column:type"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Message) TableName() string {
	return "messages"
}
