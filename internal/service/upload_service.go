package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"room-api/internal/config"
)

const maxAvatarSize = 2 * 1024 * 1024

type UploadService struct {
	client     *cos.Client
	baseURL    string
	cdnURL     string
	pathPrefix string
}

func NewUploadService(cfg config.Config) (*UploadService, error) {
	if cfg.TencentSecretID == "" || cfg.TencentSecretKey == "" || cfg.COSBaseURL == "" {
		return nil, errors.New("COS配置不能为空")
	}

	bucketURL, err := url.Parse(strings.TrimRight(cfg.COSBaseURL, "/"))
	if err != nil {
		return nil, err
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.TencentSecretID,
			SecretKey: cfg.TencentSecretKey,
		},
	})

	pathPrefix := strings.Trim(cfg.COSPathPrefix, "/")
	if pathPrefix == "" {
		pathPrefix = "room"
	}

	return &UploadService{
		client:     client,
		baseURL:    strings.TrimRight(cfg.COSBaseURL, "/"),
		cdnURL:     strings.TrimRight(cfg.COSCDNURL, "/"),
		pathPrefix: pathPrefix,
	}, nil
}

func (s *UploadService) UploadAvatar(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", errors.New("参数错误")
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxAvatarSize {
		return "", errors.New("文件过大")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	contentType, ext, ok := detectImageType(header[:n])
	if !ok {
		return "", errors.New("文件类型错误")
	}

	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("文件错误")
	}

	objectKey, err := s.avatarObjectKey(ext)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = s.client.Object.Put(ctx, objectKey, file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}

	return s.publicURL(objectKey), nil
}

func (s *UploadService) avatarObjectKey(ext string) (string, error) {
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return s.pathPrefix + "/" + hex.EncodeToString(token) + "." + ext, nil
}

func (s *UploadService) publicURL(objectKey string) string {
	base := s.baseURL
	if s.cdnURL != "" {
		base = s.cdnURL
	}
	return base + "/" + objectKey
}

func detectImageType(header []byte) (contentType string, ext string, ok bool) {
	contentType = http.DetectContentType(header)
	switch contentType {
	case "image/jpeg":
		return contentType, "jpg", true
	case "image/png":
		return contentType, "png", true
	case "image/webp":
		return contentType, "webp", true
	}

	if len(header) >= 12 &&
		string(header[0:4]) == "RIFF" &&
		string(header[8:12]) == "WEBP" {
		return "image/webp", "webp", true
	}

	return "", "", false
}
