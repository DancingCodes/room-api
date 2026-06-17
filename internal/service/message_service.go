package service

import (
	"errors"
	"strings"

	"room-api/internal/model"
	"room-api/internal/repository"
)

const (
	defaultMessageLimit = 20
	maxMessageLimit     = 50
)

type MessageService struct {
	messages *repository.MessageRepository
	rooms    *repository.RoomRepository
	users    *repository.UserRepository
}

type MessageDTO struct {
	ID              uint64 `json:"id"`
	RoomID          uint64 `json:"room_id"`
	SenderID        uint64 `json:"sender_id"`
	SenderNickname  string `json:"sender_nickname"`
	SenderAvatarURL string `json:"sender_avatar_url"`
	Type            string `json:"type"`
	Content         string `json:"content"`
	CreatedAt       string `json:"created_at"`
}

type MessageListDTO struct {
	List  []MessageDTO `json:"list"`
	Limit int          `json:"limit"`
}

func NewMessageService(messages *repository.MessageRepository, rooms *repository.RoomRepository, users *repository.UserRepository) *MessageService {
	return &MessageService{messages: messages, rooms: rooms, users: users}
}

func (s *MessageService) List(userID, roomID uint64, limit int, beforeID uint64) (*MessageListDTO, error) {
	isMember, err := s.rooms.IsMember(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("不在房间内")
	}

	limit = normalizeMessageLimit(limit)
	messages, err := s.messages.List(roomID, limit, beforeID)
	if err != nil {
		return nil, err
	}

	senders, err := s.senderMap(messages)
	if err != nil {
		return nil, err
	}

	list := make([]MessageDTO, 0, len(messages))
	for _, message := range messages {
		dto, err := s.toDTO(&message, senders)
		if err != nil {
			return nil, err
		}
		list = append(list, dto)
	}

	return &MessageListDTO{List: list, Limit: limit}, nil
}

func (s *MessageService) Create(userID, roomID uint64, content string) (*MessageDTO, error) {
	isMember, err := s.rooms.IsMember(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("不在房间内")
	}

	content = strings.TrimSpace(content)
	if runeLen(content) < 1 || runeLen(content) > 50 {
		return nil, errors.New("消息内容长度错误")
	}

	message := &model.Message{
		RoomID:   roomID,
		SenderID: userID,
		Type:     "text",
		Content:  content,
	}
	if err := s.messages.Create(message); err != nil {
		return nil, err
	}

	sender, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	dto := messageDTO(message, sender)
	return &dto, nil
}

func (s *MessageService) senderMap(messages []model.Message) (map[uint64]model.User, error) {
	senderIDs := make([]uint64, 0, len(messages))
	seen := make(map[uint64]struct{}, len(messages))
	for _, message := range messages {
		if _, ok := seen[message.SenderID]; ok {
			continue
		}
		seen[message.SenderID] = struct{}{}
		senderIDs = append(senderIDs, message.SenderID)
	}
	return s.users.FindByIDs(senderIDs)
}

func (s *MessageService) toDTO(message *model.Message, senders map[uint64]model.User) (MessageDTO, error) {
	sender, ok := senders[message.SenderID]
	if !ok {
		return MessageDTO{}, repository.ErrNotFound
	}
	return messageDTO(message, &sender), nil
}

func messageDTO(message *model.Message, sender *model.User) MessageDTO {
	return MessageDTO{
		ID:              message.ID,
		RoomID:          message.RoomID,
		SenderID:        message.SenderID,
		SenderNickname:  sender.Nickname,
		SenderAvatarURL: sender.AvatarURL,
		Type:            message.Type,
		Content:         message.Content,
		CreatedAt:       formatTime(message.CreatedAt),
	}
}

func normalizeMessageLimit(limit int) int {
	if limit < 1 {
		return defaultMessageLimit
	}
	if limit > maxMessageLimit {
		return maxMessageLimit
	}
	return limit
}
