package service

import "errors"

var ErrMailSenderNotConfigured = errors.New("邮件发送服务未配置")

type MailSender interface {
	SendVerificationCode(email string, code string) error
}

type DisabledMailSender struct{}

func (DisabledMailSender) SendVerificationCode(email string, code string) error {
	return ErrMailSenderNotConfigured
}
