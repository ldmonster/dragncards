package lfg

import (
	"errors"
	"fmt"
	"time"
)

type LfgService struct {
	repo LfgRepository
}

func NewService(repo LfgRepository) *LfgService {
	return &LfgService{repo: repo}
}

func (s *LfgService) Create(pluginID, userID, text string) (*LfgPost, error) {
	if pluginID == "" || userID == "" || text == "" {
		return nil, errors.New("plugin_id, user_id, and text are required")
	}
	post := &LfgPost{ID: fmt.Sprintf("lfg-%d", time.Now().UnixNano()), PluginID: pluginID, UserID: userID, Text: text}
	if err := s.repo.Create(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *LfgService) List(pluginID string) ([]*LfgPost, error) {
	return s.repo.List(pluginID)
}

func (s *LfgService) Delete(id string) error {
	return s.repo.Delete(id)
}
