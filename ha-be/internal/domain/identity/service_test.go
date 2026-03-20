package identity_test

import (
	"testing"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
)

func TestConfirmEmailUpdatesRepository(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	service := identity.NewService(repo)

	user, err := service.Register("test@example.com", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	token, err := service.GenerateConfirmToken(user.Email)
	if err != nil {
		t.Fatalf("generate confirm token failed: %v", err)
	}

	if err := service.ConfirmEmail(token); err != nil {
		t.Fatalf("confirm email failed: %v", err)
	}

	got, err := repo.FindByEmail(user.Email)
	if err != nil {
		t.Fatalf("find by email failed: %v", err)
	}
	if !got.Confirmed {
		t.Fatalf("expected confirmed true, got false")
	}
}

func TestResetPasswordUpdatesRepository(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	service := identity.NewService(repo)

	user, err := service.Register("test2@example.com", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	token, err := service.GenerateConfirmToken(user.Email)
	if err != nil {
		t.Fatalf("generate confirm token failed: %v", err)
	}
	if err := service.ConfirmEmail(token); err != nil {
		t.Fatalf("confirm email failed: %v", err)
	}

	resetToken, err := service.GenerateResetToken(user.Email)
	if err != nil {
		t.Fatalf("generate reset token failed: %v", err)
	}
	if err := service.ResetPassword(resetToken, "new-password"); err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	got, err := repo.FindByEmail(user.Email)
	if err != nil {
		t.Fatalf("find by email failed: %v", err)
	}
	if !identity.ComparePassword(got.PasswordHash, "new-password") {
		t.Fatalf("expected password to be updated")
	}
}

func TestConfirmTokenExpired(t *testing.T) {
	repo := persistence.NewInMemoryUserRepository()
	service := identity.NewService(repo)
	service.SetTokenTTL(1*time.Nanosecond, 1*time.Hour)

	user, err := service.Register("test3@example.com", "password123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	token, err := service.GenerateConfirmToken(user.Email)
	if err != nil {
		t.Fatalf("generate confirm token failed: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	if err := service.ConfirmEmail(token); err != identity.ErrTokenExpired {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}
