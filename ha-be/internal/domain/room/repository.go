package room

type RoomRepository interface {
	Create(room *Room) error
	List() ([]*Room, error)
	FindBySlug(slug string) (*Room, error)
	AppendAction(slug string, payload []byte) error
	ListActions(slug string) ([][]byte, error)
}
