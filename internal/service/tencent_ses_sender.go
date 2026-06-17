package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	ses "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ses/v20201002"

	"room-api/internal/config"
)

const verificationEmailSubject = "Room账号验证码"

type TencentSESMailSender struct {
	client     *ses.Client
	from       string
	templateID uint64
}

func NewTencentSESMailSender(cfg config.Config) (*TencentSESMailSender, error) {
	if cfg.TencentSecretID == "" || cfg.TencentSecretKey == "" ||
		cfg.TencentSESRegion == "" || cfg.TencentSESFrom == "" || cfg.TencentSESTemplateID == "" {
		return nil, errors.New("SES配置不能为空")
	}

	templateID, err := strconv.ParseUint(cfg.TencentSESTemplateID, 10, 64)
	if err != nil || templateID == 0 {
		return nil, errors.New("SES模板ID错误")
	}

	client, err := ses.NewClientWithSecretId(cfg.TencentSecretID, cfg.TencentSecretKey, cfg.TencentSESRegion)
	if err != nil {
		return nil, err
	}

	return &TencentSESMailSender{
		client:     client,
		from:       strings.TrimSpace(cfg.TencentSESFrom),
		templateID: templateID,
	}, nil
}

func (s *TencentSESMailSender) SendVerificationCode(email string, code string) error {
	templateData, err := json.Marshal(map[string]string{"code": code})
	if err != nil {
		return err
	}

	from := s.from
	subject := verificationEmailSubject
	destination := strings.TrimSpace(email)
	templateDataText := string(templateData)
	triggerType := uint64(1)

	request := ses.NewSendEmailRequest()
	request.FromEmailAddress = &from
	request.Subject = &subject
	request.Destination = []*string{&destination}
	request.Template = &ses.Template{
		TemplateID:   &s.templateID,
		TemplateData: &templateDataText,
	}
	request.TriggerType = &triggerType

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = s.client.SendEmailWithContext(ctx, request)
	return err
}
