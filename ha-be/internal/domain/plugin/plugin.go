package plugin

type Plugin struct {
	ID      string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Name    string `json:"name" gorm:"type:varchar(256);not null"`
	Visible bool   `json:"visible" gorm:"not null;default:false"`
	RepoURL string `json:"repo_url" gorm:"type:varchar(1024);default:null"`
}

type UserPluginPermission struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	PluginID string `json:"plugin_id" gorm:"index;type:varchar(64);not null"`
	UserID   string `json:"user_id" gorm:"index;type:varchar(64);not null"`
	Allowed  bool   `json:"allowed" gorm:"not null;default:false"`
}
