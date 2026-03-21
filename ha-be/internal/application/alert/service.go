package alert
package alert

import "github.com/ldmonster/dragncards/ha-be/internal/domain/alert"

















}	return s.domain.Create(message, level)func (s *Service) Create(message, level string) (*alert.Alert, error) {}	return s.domain.List()func (s *Service) List() ([]*alert.Alert, error) {}	return &Service{domain: domainSvc}func NewService(domainSvc *alert.AlertService) *Service {}	domain *alert.AlertServicetype Service struct {// Service wraps Alert domain service in the application layer.