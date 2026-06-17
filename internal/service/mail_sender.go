package service

import "errors"

var ErrMailSenderNotConfigured = errors.New("mail sender not configured")

type MailSender interface {
	SendVerificationCode(email string, code string) error
}

type DisabledMailSender struct{}

func (DisabledMailSender) SendVerificationCode(email string, code string) error {
	return ErrMailSenderNotConfigured
}
