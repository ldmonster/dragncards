package persistence

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user *identity.User) error {
	if user == nil || user.Email == "" {
		return errors.New("invalid user")
	}
	return r.db.Create(user).Error
}

func (r *GormUserRepository) FindByEmail(email string) (*identity.User, error) {
	if email == "" {
		return nil, ErrNoUser
	}
	var user identity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoUser
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(id string) (*identity.User, error) {
	if id == "" {
		return nil, ErrNoUser
	}
	var user identity.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoUser
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByConfirmToken(token string) (*identity.User, error) {
	if token == "" {
		return nil, ErrNoUser
	}
	var user identity.User
	if err := r.db.Where("confirm_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoUser
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByResetToken(token string) (*identity.User, error) {
	if token == "" {
		return nil, ErrNoUser
	}
	var user identity.User
	if err := r.db.Where("reset_token = ?", token).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoUser
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Update(user *identity.User) error {
	if user == nil || user.ID == "" {
		return errors.New("invalid user")
	}
	if err := r.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormUserRepository) Delete(id string) error {
	if id == "" {
		return ErrNoUser
	}
	res := r.db.Delete(&identity.User{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNoUser
	}
	return nil
}

func (r *GormUserRepository) List() ([]*identity.User, error) {
	var users []*identity.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
