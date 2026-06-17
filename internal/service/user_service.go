package service

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"room-api/internal/auth"
	"room-api/internal/model"
	"room-api/internal/repository"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{4,20}$`)

type UserService struct {
	users  *repository.UserRepository
	tokens *auth.Service
}

type UserDTO struct {
	ID            uint64  `json:"id"`
	Username      string  `json:"username"`
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

func NewUserService(users *repository.UserRepository, tokens *auth.Service) *UserService {
	return &UserService{users: users, tokens: tokens}
}

func (s *UserService) Register(username, email, password, nickname, avatarURL string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	email = normalizeEmail(email)
	nickname = strings.TrimSpace(nickname)
	avatarURL = strings.TrimSpace(avatarURL)

	if err := validateUserFields(username, email, password, nickname, avatarURL); err != nil {
		return nil, err
	}

	exists, err := s.users.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	exists, err = s.users.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	exists, err = s.users.NicknameExists(nickname, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("nickname already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		Nickname:     nickname,
		PasswordHash: string(passwordHash),
		AvatarURL:    avatarURL,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.authResult(user)
}

func (s *UserService) Login(username, password string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errors.New("invalid params")
	}

	user, err := s.users.FindByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("username or password is incorrect")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("username or password is incorrect")
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
	if runeLen(nickname) < 1 || runeLen(nickname) > 8 {
		return nil, errors.New("invalid params")
	}

	exists, err := s.users.NicknameExists(nickname, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("nickname already exists")
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
		Username:      user.Username,
		Email:         user.Email,
		Nickname:      user.Nickname,
		AvatarURL:     user.AvatarURL,
		CurrentRoomID: currentRoomID,
		CreatedAt:     formatTime(user.CreatedAt),
		UpdatedAt:     formatTime(user.UpdatedAt),
	}, nil
}

func validateUserFields(username, email, password, nickname, avatarURL string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("invalid params")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid params")
	}
	if len(password) < 6 || len(password) > 20 {
		return errors.New("invalid params")
	}
	if runeLen(nickname) < 1 || runeLen(nickname) > 8 {
		return errors.New("invalid params")
	}
	if avatarURL == "" {
		return errors.New("avatar is required")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func runeLen(value string) int {
	return len([]rune(value))
}
