package plugin

import (
	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
)

// Service is the application-level plugin service wrapper.
type Service struct {
	domain *plugin.PluginService
}

func NewService(repo plugin.PluginRepository) *Service {
	return &Service{domain: plugin.NewService(repo)}
}

func (s *Service) Create(name string, visible bool) (*plugin.Plugin, error) {
	return s.domain.Create(name, visible)
}

func (s *Service) List() ([]*plugin.Plugin, error) {
	return s.domain.List()
}

func (s *Service) ListVisible() ([]*plugin.Plugin, error) {
	return s.domain.ListVisible()
}

func (s *Service) FindByID(id string) (*plugin.Plugin, error) {
	return s.domain.FindByID(id)
}

func (s *Service) CreateCustomCard(pluginID, name, data string) (*plugin.CustomCard, error) {
	return s.domain.CreateCustomCard(pluginID, name, data)
}

func (s *Service) ListCustomCards(pluginID string) ([]*plugin.CustomCard, error) {
	return s.domain.ListCustomCards(pluginID)
}

func (s *Service) FindCustomCardByID(cardID string) (*plugin.CustomCard, error) {
	return s.domain.FindCustomCardByID(cardID)
}

func (s *Service) LoadCardsByPlugin() (map[string]map[string]any, error) {
	plugins, err := s.domain.ListVisible()
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]any{}
	for _, p := range plugins {
		cards, err := s.domain.ListCustomCards(p.ID)
		if err != nil {
			return nil, err
		}
		m := map[string]any{}
		for _, c := range cards {
			m[c.ID] = map[string]any{"id": c.ID, "name": c.Name, "data": c.Data}
		}
		out[p.ID] = m
	}
	return out, nil
}
