package lfg

import "github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"

// Service wraps LFG domain service in the application layer.
type Service struct {
	domain *lfg.LfgService
}

func NewService(domainSvc *lfg.LfgService) *Service {
	return &Service{domain: domainSvc}
}

func (s *Service) List(pluginID string) ([]*lfg.LfgPost, error) {
	return s.domain.List(pluginID)
}

func (s *Service) Create(pluginID, userID, text string) (*lfg.LfgPost, error) {
	return s.domain.Create(pluginID, userID, text)
}

func (s *Service) Delete(id string) error {
	return s.domain.Delete(id)
}
