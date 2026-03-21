package identity

import (
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
)

type Service struct {
	domainSvc *identity.IdentityService
}

func NewService(domainSvc *identity.IdentityService) *Service {
	return &Service{domainSvc: domainSvc}
}

func (s *Service) SetTokenTTL(confirm, reset time.Duration) {
	s.domainSvc.SetTokenTTL(confirm, reset)
}

func (s *Service) Register(email, password string) (*identity.User, error) {
	return s.domainSvc.Register(email, password)
}

func (s *Service) Authenticate(email, password string) (*identity.User, error) {
	return s.domainSvc.Authenticate(email, password)
}

func (s *Service) GenerateConfirmToken(email string) (string, error) {
	return s.domainSvc.GenerateConfirmToken(email)
}

func (s *Service) ConfirmEmail(token string) error {
	return s.domainSvc.ConfirmEmail(token)
}

func (s *Service) GenerateResetToken(email string) (string, error) {
	return s.domainSvc.GenerateResetToken(email)
}

func (s *Service) ResetPassword(token, newPassword string) error {
	return s.domainSvc.ResetPassword(token, newPassword)
}

func (s *Service) GetUserByID(id string) (*identity.User, error) {
	return s.domainSvc.GetUserByID(id)
}

func (s *Service) IsAdmin(userID string) (bool, error) {
	return s.domainSvc.IsAdmin(userID)
}

func (s *Service) ListUsers() ([]*identity.User, error) {
	return s.domainSvc.ListUsers()
}

func (s *Service) UpdateUser(user *identity.User) error {
	return s.domainSvc.UpdateUser(user)
}

func (s *Service) DeleteUser(id string) error {
	return s.domainSvc.DeleteUser(id)
}

func (s *Service) SetPassword(id, newPassword string) error {
	return s.domainSvc.SetPassword(id, newPassword)
}
