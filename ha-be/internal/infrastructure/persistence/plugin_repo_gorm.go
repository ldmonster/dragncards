package persistence

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	"gorm.io/gorm"
)

type GormPluginRepository struct {
	db *gorm.DB
}

func NewGormPluginRepository(db *gorm.DB) *GormPluginRepository {
	return &GormPluginRepository{db: db}
}

func (r *GormPluginRepository) Create(p *plugin.Plugin) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid plugin")
	}
	return r.db.Create(p).Error
}

func (r *GormPluginRepository) List() ([]*plugin.Plugin, error) {
	var plugins []*plugin.Plugin
	if err := r.db.Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

func (r *GormPluginRepository) ListVisible() ([]*plugin.Plugin, error) {
	var plugins []*plugin.Plugin
	if err := r.db.Where("visible = ?", true).Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

func (r *GormPluginRepository) FindByID(id string) (*plugin.Plugin, error) {
	if id == "" {
		return nil, ErrPluginNotFound
	}
	var p plugin.Plugin
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrPluginNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *GormPluginRepository) CreateCustomCard(card *plugin.CustomCard) error {
	if card == nil || card.ID == "" || card.PluginID == "" {
		return errors.New("invalid custom card")
	}
	return r.db.Create(card).Error
}

func (r *GormPluginRepository) ListCustomCards(pluginID string) ([]*plugin.CustomCard, error) {
	if pluginID == "" {
		return nil, errors.New("plugin_id required")
	}
	var cards []*plugin.CustomCard
	if err := r.db.Where("plugin_id = ?", pluginID).Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *GormPluginRepository) FindCustomCardByID(id string) (*plugin.CustomCard, error) {
	if id == "" {
		return nil, errors.New("custom card not found")
	}
	var card plugin.CustomCard
	if err := r.db.First(&card, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("custom card not found")
		}
		return nil, err
	}
	return &card, nil
}
