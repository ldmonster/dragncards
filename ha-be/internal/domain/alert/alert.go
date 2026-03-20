package alert

type Alert struct {
	ID      string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Message string `json:"message" gorm:"type:text;not null"`
	Level   string `json:"level" gorm:"type:varchar(32);not null"`
}
