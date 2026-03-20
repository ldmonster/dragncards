package identity

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	FindByConfirmToken(token string) (*User, error)
	FindByResetToken(token string) (*User, error)
	List() ([]*User, error)
	Update(user *User) error
	Delete(id string) error
}
