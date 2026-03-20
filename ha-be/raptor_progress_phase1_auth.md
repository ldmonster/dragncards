# Raptor Progress - Phase1 Auth

## Purpose
Document auth-specific implementation status.

## Status — COMPLETE (core)

### Endpoints (all in `internal/interfaces/http/handler.go` / `NewRouter`)
| Method | Path | Handler | Notes |
|---|---|---|---|
| POST | `/be/api/v1/registration` | `Register` | Creates user, sends confirm email |
| POST | `/be/api/v1/session` | `Login` | Returns `auth_token` + `renew_token` |
| DELETE | `/be/api/v1/session` | `Logout` | Revokes `auth_token` in-memory |
| POST | `/be/api/v1/session/renew` | `RenewSession` | Issues new token pair from renew token |
| GET | `/be/api/v1/confirm-email` | `ConfirmEmail` | Token via query param |
| POST | `/be/api/v1/reset-password` | `RequestPasswordReset` | Sends reset email |
| POST | `/be/api/v1/reset-password/update` | `ResetPassword` | Applies new password |
| POST | `/be/api/v1/recaptcha/verify` | `VerifyRecaptcha` | Proxies Google siteverify |

### Token layer (`internal/platform/auth/jwt.go`)
- [x] HS256 JWT via `golang-jwt/jwt v4` — `GenerateToken(subject, ttl)`, `ParseToken`, `RevokeToken`
- [x] Secret resolved: `cfg.Auth.JWTSecret` → `AUTH_TOKEN_SECRET` env → dev fallback
- [x] In-memory revocation set (thread-safe `sync.RWMutex`)
- [x] Separate access-token TTL and renew-token TTL wired from config (`cfg.Auth.AccessLifetime` minutes, `cfg.Auth.RefreshLifetime` hours)

### Recaptcha (`VerifyRecaptcha`)
- [x] Proxies Google `siteverify` API with configured `recaptchaSecret`
- [x] Returns HTTP 501 Not Implemented when secret unset (disabled in dev)

### SMTP email (`internal/infrastructure/email/smtp_mailer.go`)
- [x] `net/smtp` based `SMTPMailer` implementing `Mailer` interface
- [x] Sends confirmation and reset emails; mailer is nil-safe (operations skipped when not configured)

### Protected routes
- [x] `middleware.Auth` applied to `/be/api/lfg` and `/be/api/alerts` route groups via chi `r.Group`
- [x] WebSocket upgrade at `/be/socket` (GET) requires `?token=` query param validated via `auth.ParseToken`

## Implementation notes
- Password hashing uses SHA-256 + random salt (not bcrypt) to avoid CGO in offline builds.
- Token revocation is in-memory; refreshes survive server restart (no DB backing yet).
- JWT library stays at `v4`; `v5` upgrade has no functional impact at current scope.

## TODO
- [ ] Persistent refresh-token revocation store (Redis or DB)
- [ ] Rate-limit login/registration endpoints
- [ ] Return structured error JSON bodies (currently HTTP status codes only)
