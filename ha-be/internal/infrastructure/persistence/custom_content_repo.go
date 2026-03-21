package persistence

import (
	"github.com/ldmonster/dragncards/ha-be/internal/domain/custom_content"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
)

// CustomContentRepositoryAdapter implements custom_content.CustomContentRepository via plugin.PluginRepository.
// It stores content in the same custom_cards table used by plugin repo.
type CustomContentRepositoryAdapter struct {
	pluginRepo plugin.PluginRepository
}

func NewCustomContentRepository(pluginRepo plugin.PluginRepository) *CustomContentRepositoryAdapter {
	return &CustomContentRepositoryAdapter{pluginRepo: pluginRepo}
}

func (r *CustomContentRepositoryAdapter) Create(content *custom_content.CustomContent) error {
	if content == nil {
		return ErrInvalidPlugin
	}
	return r.pluginRepo.CreateCustomCard(&plugin.CustomCard{ID: content.ID, PluginID: content.PluginID, OwnerID: content.OwnerID, Name: content.Name, Data: content.Data})
}

func (r *CustomContentRepositoryAdapter) ListByPlugin(pluginID string) ([]*custom_content.CustomContent, error) {
	cards, err := r.pluginRepo.ListCustomCards(pluginID)
	if err != nil {
		return nil, err
	}
	out := make([]*custom_content.CustomContent, 0, len(cards))
	for _, c := range cards {
		out = append(out, &custom_content.CustomContent{ID: c.ID, PluginID: c.PluginID, OwnerID: c.OwnerID, Name: c.Name, Data: c.Data})
	}
	return out, nil
}

func (r *CustomContentRepositoryAdapter) ListByOwner(ownerID, pluginID string) ([]*custom_content.CustomContent, error) {
	cards, err := r.pluginRepo.ListCustomCardsByOwner(ownerID, pluginID)
	if err != nil {
		return nil, err
	}
	out := make([]*custom_content.CustomContent, 0, len(cards))
	for _, c := range cards {
		out = append(out, &custom_content.CustomContent{ID: c.ID, PluginID: c.PluginID, OwnerID: c.OwnerID, Name: c.Name, Data: c.Data})
	}
	return out, nil
}

func (r *CustomContentRepositoryAdapter) FindByID(id string) (*custom_content.CustomContent, error) {
	c, err := r.pluginRepo.FindCustomCardByID(id)
	if err != nil {
		return nil, err
	}
	return &custom_content.CustomContent{ID: c.ID, PluginID: c.PluginID, OwnerID: c.OwnerID, Name: c.Name, Data: c.Data}, nil
}

func (r *CustomContentRepositoryAdapter) Delete(id string) error {
	return r.pluginRepo.DeleteCustomCard(id)
}
