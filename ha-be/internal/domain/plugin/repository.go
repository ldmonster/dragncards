package plugin

type PluginRepository interface {
	Create(plugin *Plugin) error
	List() ([]*Plugin, error)
	ListVisible() ([]*Plugin, error)
	FindByID(id string) (*Plugin, error)
	Update(plugin *Plugin) error

	CreateCustomCard(card *CustomCard) error
	UpsertCustomCard(card *CustomCard) error
	ListCustomCards(pluginID string) ([]*CustomCard, error)
	FindCustomCardByID(id string) (*CustomCard, error)

	CreatePermission(permission *UserPluginPermission) error
	GetPermission(pluginID, userID string) (*UserPluginPermission, error)
	DeletePermission(pluginID, userID string) error
}
