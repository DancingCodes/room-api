package service

import (
	"errors"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"room-api/internal/model"
	"room-api/internal/repository"
)

type AppVersionDTO struct {
	VersionCode  uint64 `json:"version_code"`
	APKURL       string `json:"apk_url"`
	ReleaseNotes string `json:"release_notes"`
}

type AdminAppVersionDTO struct {
	ID           uint64 `json:"id"`
	VersionCode  uint64 `json:"version_code"`
	APKURL       string `json:"apk_url"`
	ReleaseNotes string `json:"release_notes"`
	IsPublished  bool   `json:"is_published"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type SaveAppVersionInput struct {
	VersionCode  uint64 `json:"version_code"`
	APKURL       string `json:"apk_url"`
	ReleaseNotes string `json:"release_notes"`
}

type AppVersionService struct {
	versions *repository.AppVersionRepository
}

func NewAppVersionService(versions *repository.AppVersionRepository) *AppVersionService {
	return &AppVersionService{versions: versions}
}

func (s *AppVersionService) Latest() (*AppVersionDTO, error) {
	version, err := s.versions.LatestPublished()
	if err != nil || version == nil {
		return nil, err
	}
	return &AppVersionDTO{
		VersionCode:  version.VersionCode,
		APKURL:       version.APKURL,
		ReleaseNotes: version.ReleaseNotes,
	}, nil
}

func (s *AppVersionService) LatestAdmin() (*AdminAppVersionDTO, error) {
	version, err := s.versions.LatestPublished()
	if err != nil || version == nil {
		return nil, err
	}
	dto := adminAppVersionDTO(version)
	return &dto, nil
}

func (s *AppVersionService) ListAdmin() ([]AdminAppVersionDTO, error) {
	versions, err := s.versions.List()
	if err != nil {
		return nil, err
	}
	result := make([]AdminAppVersionDTO, 0, len(versions))
	for i := range versions {
		result = append(result, adminAppVersionDTO(&versions[i]))
	}
	return result, nil
}

func (s *AppVersionService) CreateAdmin(input SaveAppVersionInput) (*AdminAppVersionDTO, error) {
	if err := validateAppVersionInput(input); err != nil {
		return nil, err
	}
	version, err := s.versions.Create(input.VersionCode, strings.TrimSpace(input.APKURL), strings.TrimSpace(input.ReleaseNotes))
	if err != nil {
		return nil, err
	}
	dto := adminAppVersionDTO(version)
	return &dto, nil
}

func (s *AppVersionService) UpdateAdmin(id uint64, input SaveAppVersionInput) (*AdminAppVersionDTO, error) {
	if err := validateAppVersionInput(input); err != nil {
		return nil, err
	}
	version, err := s.versions.UpdateDraft(id, input.VersionCode, strings.TrimSpace(input.APKURL), strings.TrimSpace(input.ReleaseNotes))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("版本不存在")
	}
	if err != nil {
		return nil, err
	}
	dto := adminAppVersionDTO(version)
	return &dto, nil
}

func (s *AppVersionService) PublishAdmin(id uint64) (*AdminAppVersionDTO, error) {
	version, err := s.versions.Publish(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("版本不存在")
	}
	if err != nil {
		return nil, err
	}
	dto := adminAppVersionDTO(version)
	return &dto, nil
}

func (s *AppVersionService) UnpublishAdmin(id uint64) error {
	err := s.versions.Unpublish(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("版本不存在")
	}
	return err
}

func validateAppVersionInput(input SaveAppVersionInput) error {
	if input.VersionCode == 0 || strings.TrimSpace(input.ReleaseNotes) == "" {
		return errors.New("参数错误")
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(input.APKURL))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return errors.New("APK 下载地址错误")
	}
	return nil
}

func adminAppVersionDTO(version *model.AppVersion) AdminAppVersionDTO {
	return AdminAppVersionDTO{
		ID:           version.ID,
		VersionCode:  version.VersionCode,
		APKURL:       version.APKURL,
		ReleaseNotes: version.ReleaseNotes,
		IsPublished:  version.IsPublished,
		CreatedAt:    formatTime(version.CreatedAt),
		UpdatedAt:    formatTime(version.UpdatedAt),
	}
}
