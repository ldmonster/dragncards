# Raptor Progress - Phase1 Identity

## Purpose
Track identity layer implementation status.

## Status — COMPLETE

### Files
- `internal/domain/identity/user.go` — `User` struct with GORM model, `ConfirmToken`, `ResetToken` fields
- `internal/domain/identity/repository.go` — `UserRepository` interface (`Create`, `FindByEmail`, `FindByID`, `Update`)
- `internal/domain/identity/service.go` — `IdentityService` with `Register`, `Authenticate`, `ConfirmEmail`, `GenerateConfirmToken`, `GenerateResetToken`, `ResetPassword`, `SetTokenTTL`
- `internal/domain/identity/service_test.go` — unit tests covering register/login/confirm/reset paths
- `internal/infrastructure/persistence/user_repo.go` — in-memory `UserRepository`
- `internal/infrastructure/persistence/user_repo_gorm.go` — GORM `UserRepository` (`Create`, `FindByEmail`, `FindByID`, `Update`) backed by PostgreSQL

### Capabilities
- [x] User registration with email uniqueness check (`ErrUserExists`)
- [x] Password hashing uses bcrypt (`golang.org/x/crypto/bcrypt`)
- [x] Login / credential validation
- [x] Confirm-email token generation and consumption with persisted `confirm_token` + expiry
- [x] Password reset token generation and update with persisted `reset_token` + expiry
- [x] Token TTL configurable via `SetTokenTTL`
- [x] GORM `AutoMigrate` for `identity.User` in `main.go`
- [x] DB-backed persistence selected at startup; falls back to in-memory when DB unavailable

## Implementation notes
- ID generation uses `time.Now().UnixNano()` string; suitable for single-node dev; upgrade to UUID for multi-node production.
- Hashing uses bcrypt with default cost in offline mode.
- Confirm/reset OTP storage changed from transient map to database fields for restart resilience.

## TODO
- [x] Switch ID generation to UUID (`github.com/google/uuid`)
