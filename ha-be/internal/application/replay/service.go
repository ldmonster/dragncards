package replay

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
)

// ReplayService performs application use cases for replays.
type ReplayService struct {
	repo replay.ReplayRepository
}

func NewReplayService(repo replay.ReplayRepository) *ReplayService {
	return &ReplayService{repo: repo}
}

func (s *ReplayService) Save(replayObj *replay.Replay) error {
	if _, err := s.repo.FindByID(replayObj.ID); err == nil {
		return errors.New("replay already exists")
	}
	return s.repo.Save(replayObj)
}

func (s *ReplayService) Update(replayObj *replay.Replay) error {
	if _, err := s.repo.FindByID(replayObj.ID); err != nil {
		return err
	}
	return s.repo.Save(replayObj)
}

func (s *ReplayService) GetByID(id string) (*replay.Replay, error) {
	return s.repo.FindByID(id)
}

func (s *ReplayService) ListByRoom(roomSlug string) ([]*replay.Replay, error) {
	return s.repo.FindByRoomSlug(roomSlug)
}

func (s *ReplayService) ListByOwner(ownerID string) ([]*replay.Replay, error) {
	return s.repo.FindByOwner(ownerID)
}

func (s *ReplayService) Delete(id string) error {
	return s.repo.Delete(id)
}
