package alert

import "github.com/ldmonster/dragncards/ha-be/internal/domain/alert"

// Service wraps Alert domain service in the application layer.
type Service struct {
	domain *alert.AlertService
}

func NewService(domainSvc *alert.AlertService) *Service {
	return &Service{domain: domainSvc}
}

func (s *Service) List() ([]*alert.Alert, error) {
	return s.domain.List()
}

func (s *Service) Create(message, level string) (*alert.Alert, error) {
	return s.domain.Create(message, level)
}
