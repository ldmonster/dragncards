package room
package room

import "github.com/ldmonster/dragncards/ha-be/internal/domain/room"

















}	return s.domain.Create(name, ownerID)func (s *Service) Create(name, ownerID string) (*room.Room, error) {}	return s.domain.List()func (s *Service) List() ([]*room.Room, error) {}	return &Service{domain: domainSvc}func NewService(domainSvc *room.RoomService) *Service {}	domain *room.RoomServicetype Service struct {// Service wraps the domain RoomService for application layer consistency.