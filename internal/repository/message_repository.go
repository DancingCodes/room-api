package repository

import (
	"sort"

	"gorm.io/gorm"

	"room-api/internal/model"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) List(roomID uint64, limit int, beforeID uint64) ([]model.Message, error) {
	query := r.db.Where("room_id = ?", roomID)
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	var messages []model.Message
	if err := query.Order("id DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	return messages, nil
}

func (r *MessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}
