package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"time"

	"room-api/internal/model"
	"room-api/internal/repository"
)

const (
	EmailPurposeRegister      = "register"
	EmailPurposeResetPassword = "reset_password"

	emailCodeTTL          = 5 * time.Minute
	emailCodeCooldown     = 60 * time.Second
	emailCodeHourlyWindow = time.Hour
	emailCodeHourlyLimit  = 5
)

type EmailCodeService struct {
	codes  *repository.EmailCodeRepository
	users  *repository.UserRepository
	sender MailSender
}

func NewEmailCodeService(codes *repository.EmailCodeRepository, users *repository.UserRepository, sender MailSender) *EmailCodeService {
	return &EmailCodeService{codes: codes, users: users, sender: sender}
}

func (s *EmailCodeService) SendRegisterCode(email string) error {
	email, err := normalizeEmailForCode(email)
	if err != nil {
		return err
	}

	exists, err := s.users.EmailExists(email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already exists")
	}

	return s.send(email, EmailPurposeRegister, nil)
}

func (s *EmailCodeService) SendResetPasswordCode(email string) error {
	email, err := normalizeEmailForCode(email)
	if err != nil {
		return err
	}

	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	return s.send(email, EmailPurposeResetPassword, &user.ID)
}

func (s *EmailCodeService) Verify(email, purpose, value string) error {
	email, err := normalizeEmailForCode(email)
	if err != nil {
		return err
	}

	now := time.Now()
	code, err := s.codes.LatestActive(email, purpose, now)
	if err != nil {
		return err
	}
	if code == nil || code.Code != value {
		return errors.New("invalid email code")
	}

	return s.codes.MarkUsed(code.ID, now)
}

func (s *EmailCodeService) send(email, purpose string, userID *uint64) error {
	now := time.Now()

	recentCount, err := s.codes.CountSince(email, purpose, now.Add(-emailCodeCooldown))
	if err != nil {
		return err
	}
	if recentCount > 0 {
		return errors.New("email code send too frequently")
	}

	hourlyCount, err := s.codes.CountSince(email, purpose, now.Add(-emailCodeHourlyWindow))
	if err != nil {
		return err
	}
	if hourlyCount >= emailCodeHourlyLimit {
		return errors.New("email code send limit exceeded")
	}

	value, err := generateEmailCode()
	if err != nil {
		return err
	}

	if err := s.sender.SendVerificationCode(email, value); err != nil {
		return err
	}

	if err := s.codes.InvalidateActive(email, purpose, now); err != nil {
		return err
	}

	return s.codes.Create(&model.EmailVerificationCode{
		UserID:    userID,
		Email:     email,
		Purpose:   purpose,
		Code:      value,
		ExpiresAt: now.Add(emailCodeTTL),
	})
}

func generateEmailCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func normalizeEmailForCode(email string) (string, error) {
	email = normalizeEmail(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", errors.New("invalid params")
	}
	return email, nil
}
