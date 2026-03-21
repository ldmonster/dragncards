package identity

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserNotFound       = errors.New("user not found")
)

type IdentityService struct {
	repo       UserRepository
	confirmTTL time.Duration
	resetTTL   time.Duration
}

func NewService(repo UserRepository) *IdentityService {
	return &IdentityService{
		repo:       repo,
		confirmTTL: 24 * time.Hour,
		resetTTL:   1 * time.Hour,
	}
}

func (s *IdentityService) SetTokenTTL(confirm, reset time.Duration) {
	if confirm > 0 {
		s.confirmTTL = confirm
	}
	if reset > 0 {
		s.resetTTL = reset
	}
}

func (s *IdentityService) Register(email, plainPassword string) (*User, error) {
	if email == "" || plainPassword == "" {
		return nil, errors.New("email and password are required")
	}

	if existing, _ := s.repo.FindByEmail(email); existing != nil {
		return nil, ErrUserExists
	}

	hash, err := hashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	user := &User{ID: randID(), Email: email, PasswordHash: hash, Confirmed: false}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *IdentityService) GenerateConfirmToken(email string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil || user == nil {
		return "", ErrUserNotFound
	}
	if user.Confirmed {
		return "", errors.New("already confirmed")
	}
	if s.confirmTTL <= 0 {
		s.confirmTTL = 24 * time.Hour
	}
	token := fmt.Sprintf("confirm-%d", time.Now().UnixNano())
	user.ConfirmToken = token
	user.ConfirmTokenExpiresAt = time.Now().Add(s.confirmTTL)
	if err := s.repo.Update(user); err != nil {
		return "", err
	}
	return token, nil
}

func (s *IdentityService) ConfirmEmail(token string) error {
	if token == "" {
		return ErrTokenInvalid
	}
	user, err := s.repo.FindByConfirmToken(token)
	if err != nil || user == nil {
		return ErrTokenInvalid
	}
	if time.Now().After(user.ConfirmTokenExpiresAt) {
		user.ConfirmToken = ""
		user.ConfirmTokenExpiresAt = time.Time{}
		_ = s.repo.Update(user)
		return ErrTokenExpired
	}
	if user.Confirmed {
		return nil
	}
	user.Confirmed = true
	user.ConfirmToken = ""
	user.ConfirmTokenExpiresAt = time.Time{}
	return s.repo.Update(user)
}

func (s *IdentityService) GenerateResetToken(email string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil || user == nil {
		return "", ErrUserNotFound
	}
	if !user.Confirmed {
		return "", errors.New("user not confirmed")
	}
	if s.resetTTL <= 0 {
		s.resetTTL = 1 * time.Hour
	}
	token := fmt.Sprintf("reset-%d", time.Now().UnixNano())
	user.ResetToken = token
	user.ResetTokenExpiresAt = time.Now().Add(s.resetTTL)
	if err := s.repo.Update(user); err != nil {
		return "", err
	}
	return token, nil
}

func (s *IdentityService) ResetPassword(token, newPassword string) error {
	if token == "" {
		return ErrTokenInvalid
	}
	user, err := s.repo.FindByResetToken(token)
	if err != nil || user == nil {
		return ErrTokenInvalid
	}
	if time.Now().After(user.ResetTokenExpiresAt) {
		user.ResetToken = ""
		user.ResetTokenExpiresAt = time.Time{}
		_ = s.repo.Update(user)
		return ErrTokenExpired
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.ResetToken = ""
	user.ResetTokenExpiresAt = time.Time{}
	return s.repo.Update(user)
}

func (s *IdentityService) GetUserByID(id string) (*User, error) {
	if id == "" {
		return nil, ErrUserNotFound
	}
	return s.repo.FindByID(id)
}

func (s *IdentityService) IsAdmin(userID string) (bool, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

func (s *IdentityService) ListUsers() ([]*User, error) {
	return s.repo.List()
}

func (s *IdentityService) UpdateUser(user *User) error {
	if user == nil || user.ID == "" {
		return errors.New("invalid user")
	}
	return s.repo.Update(user)
}

func (s *IdentityService) DeleteUser(id string) error {
	if id == "" {
		return errors.New("invalid user")
	}
	return s.repo.Delete(id)
}

func (s *IdentityService) SetPassword(id, newPassword string) error {
	if id == "" || newPassword == "" {
		return errors.New("invalid password")
	}
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	return s.repo.Update(user)
}

func (s *IdentityService) Authenticate(email, plainPassword string) (*User, error) {
	if email == "" || plainPassword == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if !comparePassword(user.PasswordHash, plainPassword) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func hashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func comparePassword(stored, password string) bool {
	if stored == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
}

func ComparePassword(stored, password string) bool {
	return comparePassword(stored, password)
}

func randID() string {
	return "u-" + uuid.NewString()
}
