package settings

import (
	"fmt"

	domain "github.com/ldmonster/dragncards/ha-be/internal/domain/settings"
)

type Service struct {
	repo domain.SettingRepository
}

func NewService(repo domain.SettingRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upsert(setting *domain.Setting) (*domain.Setting, error) {
	if setting == nil || setting.UserID == "" || setting.PluginID == "" {
		return nil, fmt.Errorf("user_id and plugin_id required")
	}
	if err := s.repo.CreateOrUpdate(setting); err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *Service) Get(userID, pluginID string) (*domain.Setting, error) {
	if userID == "" || pluginID == "" {
		return nil, fmt.Errorf("user_id and plugin_id required")
	}
	return s.repo.Get(userID, pluginID)
}

func (s *Service) List(userID string) ([]*domain.Setting, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id required")
	}
	return s.repo.ListByUser(userID)
}

func (s *Service) Delete(userID, pluginID string) error {
	if userID == "" || pluginID == "" {
		return fmt.Errorf("user_id and plugin_id required")
	}
	return s.repo.Delete(userID, pluginID)
}
