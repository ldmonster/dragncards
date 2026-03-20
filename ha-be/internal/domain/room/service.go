package room

import (
	"errors"
	"fmt"
	"time"
)

type RoomService struct {
	repo RoomRepository
}

func NewService(repo RoomRepository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) Create(name, ownerID string) (*Room, error) {
	if name == "" || ownerID == "" {
		return nil, errors.New("name and owner_id are required")
	}

	slug := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	room := &Room{ID: fmt.Sprintf("r-%d", time.Now().UnixNano()), Name: name, OwnerID: ownerID, Slug: slug}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) List() ([]*Room, error) {
	return s.repo.List()
}

func (s *RoomService) FindBySlug(slug string) (*Room, error) {
	return s.repo.FindBySlug(slug)
}

func (s *RoomService) AppendAction(slug string, payload []byte) error {
	if slug == "" || len(payload) == 0 {
		return errors.New("slug and payload are required")
	}
	return s.repo.AppendAction(slug, payload)
}

func (s *RoomService) ListActions(slug string) ([][]byte, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}
	return s.repo.ListActions(slug)
}
