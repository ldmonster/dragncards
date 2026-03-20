# Raptor Progress - Phase1 (summary)

## Purpose
Track Phase1 implementation for Identity/Auth and related endpoints.

## Status — COMPLETE

### Identity domain (`internal/domain/identity/`)
- [x] `User` aggregate with GORM model tags
- [x] `UserRepository` interface + in-memory impl (`persistence/user_repo.go`) + GORM impl (`persistence/user_repo_gorm.go`)
- [x] `IdentityService` — register, authenticate, confirm-email, reset-password, token TTL from config
- [x] Password hashing via `golang.org/x/crypto` (SHA-256 + salt; constant-time compare for offline build compatibility)
- [x] `GenerateConfirmToken` / `GenerateResetToken` stored on user record
- [x] Unit tests passing (`internal/domain/identity/service_test.go`)

### Auth platform (`internal/platform/auth/jwt.go`)
- [x] HS256 JWT via `golang-jwt/jwt v4`
- [x] `SetSecret` wired from `cfg.Auth.JWTSecret` in `main.go`; falls back to env `AUTH_TOKEN_SECRET`
- [x] In-memory token revocation (`RevokeToken` / `IsTokenRevoked`, thread-safe)

### Middleware (`internal/platform/middleware/middleware.go`)
- [x] `Auth` middleware — extracts Bearer token, calls `auth.ParseToken`, injects user-ID into context
- [x] Applied to `/be/api/lfg` and `/be/api/alerts` route groups

### Email (`internal/infrastructure/email/smtp_mailer.go`)
- [x] `Mailer` interface + SMTP implementation via `net/smtp`
- [x] `SendConfirmation` and `SendReset` called from identity handlers; nil-safe (skipped when unconfigured)

## Implementation notes
- JWT library is `v4` (not `v5`); upgrade deferred — no functional difference at this scope.
- Token revocation is in-memory; server restart clears revoked list (acceptable for current scope).
- JWT secret resolved at runtime from config/env; dev fallback is logged as warning.

## TODO
- [ ] Persistent refresh-token revocation store (DB or Redis backed)
- [ ] Rate-limiting on `/session` and `/registration` endpoints
