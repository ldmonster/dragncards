package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
)

var ErrNoUser = errors.New("user not found")

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*identity.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{users: map[string]*identity.User{}}
}

func (r *InMemoryUserRepository) Create(user *identity.User) error {
	if user == nil || user.Email == "" {
		return errors.New("invalid user")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == user.Email {
			return identity.ErrUserExists
		}
	}

	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*identity.User, error) {
	if email == "" {
		return nil, ErrNoUser
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, ErrNoUser
}

func (r *InMemoryUserRepository) FindByID(id string) (*identity.User, error) {
	if id == "" {
		return nil, ErrNoUser
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, ErrNoUser
	}
	return u, nil
}

func (r *InMemoryUserRepository) FindByConfirmToken(token string) (*identity.User, error) {
	if token == "" {
		return nil, ErrNoUser
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.ConfirmToken == token {
			return u, nil
		}
	}

	return nil, ErrNoUser
}

func (r *InMemoryUserRepository) FindByResetToken(token string) (*identity.User, error) {
	if token == "" {
		return nil, ErrNoUser
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.ResetToken == token {
			return u, nil
		}
	}

	return nil, ErrNoUser
}

func (r *InMemoryUserRepository) Update(user *identity.User) error {
	if user == nil || user.ID == "" {
		return errors.New("invalid user")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.users[user.ID]; !ok {
		return ErrNoUser
	} else {
		existing.Email = user.Email
		existing.PasswordHash = user.PasswordHash
		existing.Confirmed = user.Confirmed
		existing.ConfirmToken = user.ConfirmToken
		existing.ConfirmTokenExpiresAt = user.ConfirmTokenExpiresAt
		existing.ResetToken = user.ResetToken
		existing.ResetTokenExpiresAt = user.ResetTokenExpiresAt
	}

	return nil
}

func (r *InMemoryUserRepository) Delete(id string) error {
	if id == "" {
		return ErrNoUser
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrNoUser
	}
	delete(r.users, id)
	return nil
}

func (r *InMemoryUserRepository) List() ([]*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*identity.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}
