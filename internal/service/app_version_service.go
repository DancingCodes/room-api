package service

import "room-api/internal/repository"

type AppVersionDTO struct {
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
