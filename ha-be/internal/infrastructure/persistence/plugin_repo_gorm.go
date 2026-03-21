package persistence

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
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

func (r *GormPluginRepository) ListCustomCardsByOwner(ownerID, pluginID string) ([]*plugin.CustomCard, error) {
	if ownerID == "" || pluginID == "" {
		return nil, errors.New("owner_id and plugin_id required")
	}
	var cards []*plugin.CustomCard
	if err := r.db.Where("plugin_id = ? AND owner_id = ?", pluginID, ownerID).Find(&cards).Error; err != nil {
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

func (r *GormPluginRepository) CreatePermission(permission *plugin.UserPluginPermission) error {
	if permission == nil || permission.PluginID == "" || permission.UserID == "" {
		return errors.New("invalid permission")
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plugin_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"allowed"}),
	}).Create(permission).Error
}

func (r *GormPluginRepository) GetPermission(pluginID, userID string) (*plugin.UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, errors.New("permission lookup requires plugin_id and user_id")
	}
	var perm plugin.UserPluginPermission
	if err := r.db.First(&perm, "plugin_id = ? AND user_id = ?", pluginID, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &plugin.UserPluginPermission{PluginID: pluginID, UserID: userID, Allowed: false}, nil
		}
		return nil, err
	}
	return &perm, nil
}

func (r *GormPluginRepository) DeletePermission(pluginID, userID string) error {
	if pluginID == "" || userID == "" {
		return errors.New("permission delete requires plugin_id and user_id")
	}
	return r.db.Where("plugin_id = ? AND user_id = ?", pluginID, userID).Delete(&plugin.UserPluginPermission{}).Error
}

func (r *GormPluginRepository) Update(p *plugin.Plugin) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid plugin")
	}
	return r.db.Save(p).Error
}

func (r *GormPluginRepository) DeleteCustomCard(id string) error {
	if id == "" {
		return errors.New("custom card id required")
	}
	res := r.db.Delete(&plugin.CustomCard{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("custom card not found")
	}
	return nil
}

func (r *GormPluginRepository) UpsertCustomCard(card *plugin.CustomCard) error {
	if card == nil || card.ID == "" || card.PluginID == "" {
		return errors.New("invalid custom card")
	}
	return r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "data", "plugin_id"})}).Create(card).Error
}
