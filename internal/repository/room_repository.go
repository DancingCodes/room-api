package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"room-api/internal/model"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) List(page, pageSize int) ([]model.Room, int64, error) {
	var total int64
	if err := r.db.Model(&model.Room{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rooms []model.Room
	offset := (page - 1) * pageSize
	err := r.db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&rooms).Error
	if err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
}

func (r *RoomRepository) Create(owner *model.User, maxMembers uint8) (*model.Room, []model.RoomMember, error) {
	var room model.Room
	var members []model.RoomMember

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.RoomMember{}).Where("user_id = ?", owner.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("user already in room")
		}

		room = model.Room{
			Name:       owner.Nickname + " 的房间",
			OwnerID:    owner.ID,
			MaxMembers: maxMembers,
		}
		if err := tx.Create(&room).Error; err != nil {
			return err
		}

		now := time.Now()
		member := model.RoomMember{
			RoomID:    room.ID,
			UserID:    owner.ID,
			IsOwner:   true,
			MicStatus: "off",
			JoinedAt:  now,
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		members = []model.RoomMember{member}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &room, members, nil
}

func (r *RoomRepository) Detail(roomID uint64) (*model.Room, []model.RoomMember, error) {
	room, err := r.FindRoom(roomID)
	if err != nil {
		return nil, nil, err
	}

	members, err := r.ListMembers(roomID)
	if err != nil {
		return nil, nil, err
	}
	return room, members, nil
}

func (r *RoomRepository) Join(roomID, userID uint64) (*model.Room, []model.RoomMember, error) {
	var room model.Room

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.RoomMember{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("user already in room")
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&room, "id = ?", roomID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("room not found")
			}
			return err
		}

		if err := tx.Model(&model.RoomMember{}).Where("room_id = ?", roomID).Count(&count).Error; err != nil {
			return err
		}
		if count >= int64(room.MaxMembers) {
			return errors.New("room is full")
		}

		member := model.RoomMember{
			RoomID:    roomID,
			UserID:    userID,
			IsOwner:   false,
			MicStatus: "off",
			JoinedAt:  time.Now(),
		}
		return tx.Create(&member).Error
	})
	if err != nil {
		return nil, nil, err
	}

	members, err := r.ListMembers(roomID)
	if err != nil {
		return nil, nil, err
	}
	return &room, members, nil
}

type LeaveResult struct {
	Left              bool
	DeletedRoom       bool
	OwnerChanged      bool
	OldOwnerUserID    uint64
	NewOwnerUserID    uint64
	CurrentMemberSize int64
	RemainingMembers  []model.RoomMember
}

func (r *RoomRepository) Leave(roomID, userID uint64) (*LeaveResult, error) {
	result := &LeaveResult{}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var member model.RoomMember
		err := tx.First(&member, "room_id = ? AND user_id = ?", roomID, userID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		result.Left = true
		result.OldOwnerUserID = member.UserID

		if err := tx.Delete(&model.RoomMember{}, "id = ?", member.ID).Error; err != nil {
			return err
		}

		var remaining []model.RoomMember
		if err := tx.Order("joined_at ASC").Find(&remaining, "room_id = ?", roomID).Error; err != nil {
			return err
		}
		result.RemainingMembers = remaining
		result.CurrentMemberSize = int64(len(remaining))

		if len(remaining) == 0 {
			if err := tx.Delete(&model.Message{}, "room_id = ?", roomID).Error; err != nil {
				return err
			}
			if err := tx.Delete(&model.Room{}, "id = ?", roomID).Error; err != nil {
				return err
			}
			result.DeletedRoom = true
			return nil
		}

		if member.IsOwner {
			newOwner := remaining[0]
			result.OwnerChanged = true
			result.NewOwnerUserID = newOwner.UserID
			if err := tx.Model(&model.Room{}).Where("id = ?", roomID).Update("owner_id", newOwner.UserID).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.RoomMember{}).Where("room_id = ?", roomID).Update("is_owner", false).Error; err != nil {
				return err
			}
			return tx.Model(&model.RoomMember{}).Where("id = ?", newOwner.ID).Update("is_owner", true).Error
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RoomRepository) UpdateMicStatus(roomID, userID uint64, micStatus string) (*model.RoomMember, error) {
	var member model.RoomMember
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&member, "room_id = ? AND user_id = ?", roomID, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("not room member")
			}
			return err
		}

		if err := tx.Model(&model.RoomMember{}).Where("id = ?", member.ID).Update("mic_status", micStatus).Error; err != nil {
			return err
		}
		member.MicStatus = micStatus
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *RoomRepository) FindRoom(roomID uint64) (*model.Room, error) {
	var room model.Room
	if err := r.db.First(&room, "id = ?", roomID).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) IsMember(roomID, userID uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&model.RoomMember{}).Where("room_id = ? AND user_id = ?", roomID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RoomRepository) ListMembers(roomID uint64) ([]model.RoomMember, error) {
	var members []model.RoomMember
	if err := r.db.Order("joined_at ASC").Find(&members, "room_id = ?", roomID).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *RoomRepository) CountMembers(roomID uint64) (int64, error) {
	var count int64
	if err := r.db.Model(&model.RoomMember{}).Where("room_id = ?", roomID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
