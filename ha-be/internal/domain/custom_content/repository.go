package custom_content

type CustomContentRepository interface {
	Create(content *CustomContent) error
	ListByPlugin(pluginID string) ([]*CustomContent, error)
	ListByOwner(ownerID, pluginID string) ([]*CustomContent, error)
	FindByID(id string) (*CustomContent, error)
	Delete(id string) error
}
