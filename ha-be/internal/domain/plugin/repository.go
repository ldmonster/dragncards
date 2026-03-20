package plugin

type PluginRepository interface {
	Create(plugin *Plugin) error
	List() ([]*Plugin, error)
	ListVisible() ([]*Plugin, error)
	FindByID(id string) (*Plugin, error)

	CreateCustomCard(card *CustomCard) error
	ListCustomCards(pluginID string) ([]*CustomCard, error)
	FindCustomCardByID(id string) (*CustomCard, error)
}
