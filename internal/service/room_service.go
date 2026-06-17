package service

import (
	"errors"
	"math"

	"gorm.io/gorm"

	"room-api/internal/model"
	"room-api/internal/repository"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 50
)

type RoomService struct {
	rooms *repository.RoomRepository
	users *repository.UserRepository
}

type RoomDTO struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	OwnerID        uint64 `json:"owner_id"`
	CurrentMembers int64  `json:"current_members"`
	MaxMembers     uint8  `json:"max_members"`
	CreatedAt      string `json:"created_at"`
}

type RoomMemberDTO struct {
	UserID    uint64 `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	IsOwner   bool   `json:"is_owner"`
	MicStatus string `json:"mic_status"`
	JoinedAt  string `json:"joined_at"`
}

type RoomDetailDTO struct {
	Room    RoomDTO         `json:"room"`
	Members []RoomMemberDTO `json:"members"`
}

type RoomListDTO struct {
	List     []RoomDTO `json:"list"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type LeaveResultDTO struct {
	Left              bool
	DeletedRoom       bool
	OwnerChanged      bool
	NewOwnerUserID    uint64
	CurrentMemberSize int64
	RemainingMembers  []RoomMemberDTO
}

func NewRoomService(rooms *repository.RoomRepository, users *repository.UserRepository) *RoomService {
	return &RoomService{rooms: rooms, users: users}
}

func (s *RoomService) List(page, pageSize int) (*RoomListDTO, error) {
	page, pageSize = normalizePage(page, pageSize)

	rooms, total, err := s.rooms.List(page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]RoomDTO, 0, len(rooms))
	for _, item := range rooms {
		list = append(list, roomDTO(&item.Room, item.CurrentCount))
	}

	return &RoomListDTO{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *RoomService) Create(userID uint64, maxMembers uint8) (*RoomDetailDTO, error) {
	if maxMembers != 2 && maxMembers != 8 {
		return nil, errors.New("房间人数只能是2人或8人")
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}

	room, members, err := s.rooms.Create(user, maxMembers)
	if err != nil {
		return nil, err
	}
	return s.detailDTO(room, members)
}

func (s *RoomService) Detail(userID, roomID uint64) (*RoomDetailDTO, error) {
	isMember, err := s.rooms.IsMember(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("不在房间内")
	}

	room, members, err := s.rooms.Detail(roomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("房间不存在")
		}
		return nil, err
	}
	return s.detailDTO(room, members)
}

func (s *RoomService) Join(userID, roomID uint64) (*RoomDetailDTO, error) {
	room, members, err := s.rooms.Join(roomID, userID)
	if err != nil {
		return nil, err
	}
	return s.detailDTO(room, members)
}

func (s *RoomService) Leave(userID, roomID uint64) (*LeaveResultDTO, error) {
	result, err := s.rooms.Leave(roomID, userID)
	if err != nil {
		return nil, err
	}

	memberDTOs, err := s.memberDTOs(result.RemainingMembers)
	if err != nil {
		return nil, err
	}

	return &LeaveResultDTO{
		Left:              result.Left,
		DeletedRoom:       result.DeletedRoom,
		OwnerChanged:      result.OwnerChanged,
		NewOwnerUserID:    result.NewOwnerUserID,
		CurrentMemberSize: result.CurrentMemberSize,
		RemainingMembers:  memberDTOs,
	}, nil
}

func (s *RoomService) UpdateMicStatus(userID, roomID uint64, micStatus string) (*RoomMemberDTO, error) {
	if micStatus != "on" && micStatus != "off" {
		return nil, errors.New("麦克风状态错误")
	}

	member, err := s.rooms.UpdateMicStatus(roomID, userID, micStatus)
	if err != nil {
		return nil, err
	}

	members, err := s.memberDTOs([]model.RoomMember{*member})
	if err != nil {
		return nil, err
	}
	return &members[0], nil
}

func (s *RoomService) detailDTO(room *model.Room, members []model.RoomMember) (*RoomDetailDTO, error) {
	memberDTOs, err := s.memberDTOs(members)
	if err != nil {
		return nil, err
	}

	return &RoomDetailDTO{
		Room:    roomDTO(room, int64(len(members))),
		Members: memberDTOs,
	}, nil
}

func roomDTO(room *model.Room, currentMembers int64) RoomDTO {
	return RoomDTO{
		ID:             room.ID,
		Name:           room.Name,
		OwnerID:        room.OwnerID,
		CurrentMembers: currentMembers,
		MaxMembers:     room.MaxMembers,
		CreatedAt:      formatTime(room.CreatedAt),
	}
}

func (s *RoomService) memberDTOs(members []model.RoomMember) ([]RoomMemberDTO, error) {
	userIDs := make([]uint64, 0, len(members))
	seen := make(map[uint64]struct{}, len(members))
	for _, member := range members {
		if _, ok := seen[member.UserID]; ok {
			continue
		}
		seen[member.UserID] = struct{}{}
		userIDs = append(userIDs, member.UserID)
	}

	users, err := s.users.FindByIDs(userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]RoomMemberDTO, 0, len(members))
	for _, member := range members {
		user, ok := users[member.UserID]
		if !ok {
			return nil, gorm.ErrRecordNotFound
		}
		result = append(result, RoomMemberDTO{
			UserID:    member.UserID,
			Nickname:  user.Nickname,
			AvatarURL: user.AvatarURL,
			IsOwner:   member.IsOwner,
			MicStatus: member.MicStatus,
			JoinedAt:  formatTime(member.JoinedAt),
		})
	}
	return result, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page > math.MaxInt/pageSize {
		page = defaultPage
	}
	return page, pageSize
}
