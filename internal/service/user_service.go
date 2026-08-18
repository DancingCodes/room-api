package service

import (
	"errors"
	"net/mail"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"room-api/internal/auth"
	"room-api/internal/model"
	"room-api/internal/repository"
)

type UserService struct {
	users  *repository.UserRepository
	tokens *auth.Service
	codes  *EmailCodeService
}

type UserDTO struct {
	ID            uint64  `json:"id"`
	Email         string  `json:"email"`
	Nickname      string  `json:"nickname"`
	AvatarURL     string  `json:"avatar_url"`
	CurrentRoomID *uint64 `json:"current_room_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type AuthResult struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

func NewUserService(users *repository.UserRepository, tokens *auth.Service, codes *EmailCodeService) *UserService {
	return &UserService{users: users, tokens: tokens, codes: codes}
}

func (s *UserService) EmailLogin(email, emailCode string) (*AuthResult, error) {
	email = normalizeEmail(email)
	emailCode = strings.TrimSpace(emailCode)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("参数错误")
	}
	if emailCode == "" {
		return nil, errors.New("验证码错误")
	}

	if err := s.codes.Verify(email, EmailPurposeLogin, emailCode); err != nil {
		return nil, err
	}

	user, err := s.users.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user, err = s.createEmailUser(email)
	}
	if err != nil {
		return nil, err
	}

	return s.authResult(user)
}

func (s *UserService) Me(userID uint64) (*UserDTO, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	dto, err := s.toDTO(user)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *UserService) UpdateNickname(userID uint64, nickname string) (*UserDTO, error) {
	nickname = strings.TrimSpace(nickname)
	if !validNickname(nickname) {
		return nil, errors.New("参数错误")
	}

	exists, err := s.users.NicknameExists(nickname, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("昵称已存在")
	}

	user, err := s.users.UpdateNickname(userID, nickname)
	if err != nil {
		return nil, err
	}

	dto, err := s.toDTO(user)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *UserService) UpdateAvatar(userID uint64, avatarURL string) (*UserDTO, error) {
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return nil, errors.New("头像不能为空")
	}

	user, err := s.users.UpdateAvatar(userID, avatarURL)
	if err != nil {
		return nil, err
	}

	dto, err := s.toDTO(user)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *UserService) createEmailUser(email string) (*model.User, error) {
	nickname, err := s.availableNickname(email)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:     email,
		Nickname:  nickname,
		AvatarURL: defaultAvatarURL(email),
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) availableNickname(email string) (string, error) {
	base := nicknameBase(email)
	for i := 0; i < 10000; i++ {
		candidate := base
		if i > 0 {
			suffix := decimalString(i)
			prefixLen := 8 - runeLen(suffix)
			if prefixLen < 1 {
				prefixLen = 1
			}
			candidate = takeRunes(base, prefixLen) + suffix
		}

		exists, err := s.users.NicknameExists(candidate, 0)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("昵称已存在")
}

func (s *UserService) authResult(user *model.User) (*AuthResult, error) {
	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	dto, err := s.toDTO(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: dto}, nil
}

func (s *UserService) toDTO(user *model.User) (UserDTO, error) {
	currentRoomID, err := s.users.CurrentRoomID(user.ID)
	if err != nil {
		return UserDTO{}, err
	}

	return UserDTO{
		ID:            user.ID,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		CurrentRoomID: currentRoomID,
		CreatedAt:     formatTime(user.CreatedAt),
		UpdatedAt:     formatTime(user.UpdatedAt),
	}, nil
}

func defaultAvatarURL(email string) string {
	seed := url.QueryEscape(email)
	return "https://api.dicebear.com/9.x/initials/png?seed=" + seed + "&size=128"
}
func validNickname(nickname string) bool {
	return runeLen(nickname) >= 1 && runeLen(nickname) <= 8
}

func nicknameBase(email string) string {
	localPart := email
	if index := strings.Index(email, "@"); index >= 0 {
		localPart = email[:index]
	}
	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		localPart = "用户"
	}
	return takeRunes(localPart, 8)
}

func takeRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func decimalString(value int) string {
	if value == 0 {
		return "0"
	}

	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func runeLen(value string) int {
	return len([]rune(value))
}
