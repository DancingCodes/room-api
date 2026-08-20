package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	rtctokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder2"
)

const agoraTokenLifetime = time.Hour

type AgoraTokenService struct {
	appID       string
	certificate string
}

type AgoraTokenDTO struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func NewAgoraTokenService(appID, certificate string) *AgoraTokenService {
	return &AgoraTokenService{appID: appID, certificate: certificate}
}

func (s *AgoraTokenService) Issue(roomID, userID uint64) (*AgoraTokenDTO, error) {
	if s.appID == "" || s.certificate == "" {
		return nil, errors.New("声网服务未配置")
	}
	if userID == 0 || userID > math.MaxUint32 {
		return nil, errors.New("用户 ID 不支持语音")
	}

	token, err := rtctokenbuilder.BuildTokenWithUid(
		s.appID,
		s.certificate,
		fmt.Sprintf("room-%d", roomID),
		uint32(userID),
		rtctokenbuilder.RolePublisher,
		uint32(agoraTokenLifetime.Seconds()),
		uint32(agoraTokenLifetime.Seconds()),
	)
	if err != nil {
		return nil, err
	}

	return &AgoraTokenDTO{
		Token:     token,
		ExpiresAt: time.Now().Add(agoraTokenLifetime).UTC().Format(time.RFC3339),
	}, nil
}
