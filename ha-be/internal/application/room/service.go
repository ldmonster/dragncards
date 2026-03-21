package room

import "github.com/ldmonster/dragncards/ha-be/internal/domain/room"

// Service wraps the domain RoomService for application layer consistency.
type Service struct {
	domain *room.RoomService
}

func NewService(domainSvc *room.RoomService) *Service {
	return &Service{domain: domainSvc}
}

func (s *Service) List() ([]*room.Room, error) {
	return s.domain.List()
}

func (s *Service) Create(name, ownerID string) (*room.Room, error) {
	return s.domain.Create(name, ownerID)
}
