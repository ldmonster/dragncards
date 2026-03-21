package persistence

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
)

type GormLfgRepository struct {
	db *gorm.DB
}

func NewGormLfgRepository(db *gorm.DB) *GormLfgRepository {
	return &GormLfgRepository{db: db}
}

func (r *GormLfgRepository) Create(post *lfg.LfgPost) error {
	if post == nil || post.ID == "" {
		return errors.New("invalid lfg post")
	}
	return r.db.Create(post).Error
}

func (r *GormLfgRepository) List(pluginID string) ([]*lfg.LfgPost, error) {
	var posts []*lfg.LfgPost
	query := r.db
	if pluginID != "" {
		query = query.Where("plugin_id = ?", pluginID)
	}
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *GormLfgRepository) Delete(id string) error {
	if id == "" {
		return ErrLfgNotFound
	}
	result := r.db.Delete(&lfg.LfgPost{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrLfgNotFound
	}
	return nil
}
