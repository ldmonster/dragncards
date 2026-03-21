package custom_content

import (
	"fmt"
	"time"
)

type CustomContentService struct {
	repo CustomContentRepository
}

func NewService(repo CustomContentRepository) *CustomContentService {
	return &CustomContentService{repo: repo}
}

func (s *CustomContentService) ListByPlugin(pluginID string) ([]*CustomContent, error) {
	if pluginID == "" {
		return nil, fmt.Errorf("plugin_id required")
	}
	return s.repo.ListByPlugin(pluginID)
}

func (s *CustomContentService) ListByOwner(ownerID, pluginID string) ([]*CustomContent, error) {
	if ownerID == "" || pluginID == "" {
		return nil, fmt.Errorf("owner_id and plugin_id required")
	}
	return s.repo.ListByOwner(ownerID, pluginID)
}

func (s *CustomContentService) Create(pluginID, ownerID, name, data string) (*CustomContent, error) {
	if pluginID == "" || ownerID == "" || name == "" {
		return nil, fmt.Errorf("plugin_id, owner_id and name required")
	}
	content := &CustomContent{ID: fmt.Sprintf("cc-%d", time.Now().UnixNano()), PluginID: pluginID, OwnerID: ownerID, Name: name, Data: data}
	if err := s.repo.Create(content); err != nil {
		return nil, err
	}
	return content, nil
}

func (s *CustomContentService) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	return s.repo.Delete(id)
}

func (s *CustomContentService) GetByID(id string) (*CustomContent, error) {
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	return s.repo.FindByID(id)
}
