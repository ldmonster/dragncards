package alert

import (
	"fmt"
	"time"
)

type AlertService struct {
	repo AlertRepository
}

func NewService(repo AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

func (s *AlertService) Create(message, level string) (*Alert, error) {
	if message == "" || level == "" {
		return nil, fmt.Errorf("message and level are required")
	}
	al := &Alert{ID: fmt.Sprintf("a-%d", time.Now().UnixNano()), Message: message, Level: level}
	if err := s.repo.Create(al); err != nil {
		return nil, err
	}
	return al, nil
}

func (s *AlertService) List() ([]*Alert, error) {
	return s.repo.List()
}
