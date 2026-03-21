package plugin

import (
	"fmt"
	"time"
)

type PluginService struct {
	repo PluginRepository
}

func NewService(repo PluginRepository) *PluginService {
	return &PluginService{repo: repo}
}

func (s *PluginService) Create(name string, visible bool) (*Plugin, error) {
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	plug := &Plugin{ID: fmt.Sprintf("p-%d", time.Now().UnixNano()), Name: name, Visible: visible}
	if err := s.repo.Create(plug); err != nil {
		return nil, err
	}
	return plug, nil
}

func (s *PluginService) List() ([]*Plugin, error) {
	return s.repo.List()
}

func (s *PluginService) ListVisible() ([]*Plugin, error) {
	return s.repo.ListVisible()
}

func (s *PluginService) FindByID(id string) (*Plugin, error) {
	return s.repo.FindByID(id)
}

func (s *PluginService) CreateCustomCard(pluginID, name, data string) (*CustomCard, error) {
	if pluginID == "" || name == "" {
		return nil, fmt.Errorf("plugin_id and name required")
	}
	_, err := s.repo.FindByID(pluginID)
	if err != nil {
		return nil, err
	}
	card := &CustomCard{ID: fmt.Sprintf("cc-%d", time.Now().UnixNano()), PluginID: pluginID, Name: name, Data: data}
	if err := s.repo.CreateCustomCard(card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *PluginService) ListCustomCards(pluginID string) ([]*CustomCard, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin_id required")
	}
	return s.repo.ListCustomCards(pluginID)
}

func (s *PluginService) FindCustomCardByID(id string) (*CustomCard, error) {
	if id == "" {
		return nil, fmt.Errorf("card_id required")
	}
	return s.repo.FindCustomCardByID(id)
}

func (s *PluginService) SetUserPluginPermission(pluginID, userID string, allowed bool) (*UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, fmt.Errorf("plugin_id and user_id required")
	}
	_, err := s.repo.FindByID(pluginID)
	if err != nil {
		return nil, err
	}
	perm := &UserPluginPermission{ID: fmt.Sprintf("upp-%d", time.Now().UnixNano()), PluginID: pluginID, UserID: userID, Allowed: allowed}
	if err := s.repo.CreatePermission(perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *PluginService) GetUserPluginPermission(pluginID, userID string) (*UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, fmt.Errorf("plugin_id and user_id required")
	}
	return s.repo.GetPermission(pluginID, userID)
}

func (s *PluginService) HasPluginAccess(userID, pluginID string) (bool, error) {
	if userID == "" || pluginID == "" {
		return false, fmt.Errorf("user_id and plugin_id required")
	}
	perm, err := s.repo.GetPermission(pluginID, userID)
	if err != nil {
		return false, err
	}
	return perm != nil && perm.Allowed, nil
}

func (s *PluginService) DeleteUserPluginPermission(pluginID, userID string) error {
	if pluginID == "" || userID == "" {
		return fmt.Errorf("plugin_id and user_id required")
	}
	return s.repo.DeletePermission(pluginID, userID)
}
