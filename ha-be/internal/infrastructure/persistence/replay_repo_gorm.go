package persistence

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
)

type GormReplayRepository struct {
	db *gorm.DB
}

func NewGormReplayRepository(db *gorm.DB) *GormReplayRepository {
	return &GormReplayRepository{db: db}
}

func (r *GormReplayRepository) Save(rp *replay.Replay) error {
	if rp == nil || rp.ID == "" {
		return errors.New("invalid replay")
	}
	return r.db.Save(rp).Error
}

func (r *GormReplayRepository) FindByID(id string) (*replay.Replay, error) {
	if id == "" {
		return nil, ErrReplayNotFound
	}
	var rp replay.Replay
	if err := r.db.First(&rp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReplayNotFound
		}
		return nil, err
	}
	return &rp, nil
}

func (r *GormReplayRepository) FindByRoomSlug(roomSlug string) ([]*replay.Replay, error) {
	if roomSlug == "" {
		return nil, errors.New("room_slug required")
	}
	var list []*replay.Replay
	if err := r.db.Where("room_slug = ?", roomSlug).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormReplayRepository) FindByOwner(ownerID string) ([]*replay.Replay, error) {
	if ownerID == "" {
		return nil, errors.New("owner_id required")
	}
	var list []*replay.Replay
	if err := r.db.Where("owner_id = ?", ownerID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormReplayRepository) Delete(id string) error {
	if id == "" {
		return ErrReplayNotFound
	}
	res := r.db.Delete(&replay.Replay{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrReplayNotFound
	}
	return nil
}
