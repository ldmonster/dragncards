package room

type Room struct {
	ID      string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Slug    string `json:"slug" gorm:"uniqueIndex;type:varchar(128);not null"`
	Name    string `json:"name" gorm:"type:varchar(256);not null"`
	OwnerID string `json:"owner_id" gorm:"type:varchar(64);not null"`
}

type RoomAction struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug    string `json:"slug" gorm:"index;type:varchar(128);not null"`
	Payload []byte `json:"payload"`
	Created int64  `json:"created" gorm:"autoCreateTime:nano"`
}
