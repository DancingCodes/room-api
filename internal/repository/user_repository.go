package repository

import (
	"errors"

	"gorm.io/gorm"

	"room-api/internal/model"
)

var ErrNotFound = gorm.ErrRecordNotFound

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	return r.exists("username = ?", username)
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	return r.exists("email = ?", email)
}

func (r *UserRepository) NicknameExists(nickname string, excludeUserID uint64) (bool, error) {
	query := r.db.Model(&model.User{}).Where("nickname = ?", nickname)
	if excludeUserID > 0 {
		query = query.Where("id <> ?", excludeUserID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) UpdateNickname(userID uint64, nickname string) (*model.User, error) {
	if err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("nickname", nickname).Error; err != nil {
		return nil, err
	}
	return r.FindByID(userID)
}

func (r *UserRepository) UpdateAvatar(userID uint64, avatarURL string) (*model.User, error) {
	if err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error; err != nil {
		return nil, err
	}
	return r.FindByID(userID)
}

func (r *UserRepository) UpdatePassword(userID uint64, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

func (r *UserRepository) CurrentRoomID(userID uint64) (*uint64, error) {
	var member model.RoomMember
	err := r.db.Select("room_id").First(&member, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &member.RoomID, nil
}

func (r *UserRepository) exists(query string, args ...any) (bool, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Where(query, args...).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
