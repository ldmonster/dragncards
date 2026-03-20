package persistence

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"gorm.io/gorm"
)

type GormRoomRepository struct {
	db *gorm.DB
}

func NewGormRoomRepository(db *gorm.DB) *GormRoomRepository {
	return &GormRoomRepository{db: db}
}

func (r *GormRoomRepository) Create(roomItem *room.Room) error {
	if roomItem == nil || roomItem.Slug == "" {
		return errors.New("invalid room")
	}
	return r.db.Create(roomItem).Error
}

func (r *GormRoomRepository) List() ([]*room.Room, error) {
	var rooms []*room.Room
	if err := r.db.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *GormRoomRepository) FindBySlug(slug string) (*room.Room, error) {
	if slug == "" {
		return nil, ErrRoomNotFound
	}
	var roomItem room.Room
	if err := r.db.Where("slug = ?", slug).First(&roomItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	return &roomItem, nil
}

func (r *GormRoomRepository) AppendAction(slug string, payload []byte) error {
	if slug == "" || len(payload) == 0 {
		return errors.New("invalid action")
	}

	var roomItem room.Room
	if err := r.db.Where("slug = ?", slug).First(&roomItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRoomNotFound
		}
		return err
	}

	action := room.RoomAction{Slug: slug, Payload: payload}
	return r.db.Create(&action).Error
}

func (r *GormRoomRepository) ListActions(slug string) ([][]byte, error) {
	if slug == "" {
		return nil, ErrRoomNotFound
	}

	var roomItem room.Room
	if err := r.db.Where("slug = ?", slug).First(&roomItem).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	var actions []room.RoomAction
	if err := r.db.Where("slug = ?", slug).Order("id asc").Find(&actions).Error; err != nil {
		return nil, err
	}

	payloads := make([][]byte, len(actions))
	for i, a := range actions {
		payloads[i] = a.Payload
	}
	return payloads, nil
}
