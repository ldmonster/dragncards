# Raptor Review - phase1 identity

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 1 section, §2 Identity context, §7 Auth Design)
- `raptor_progress_phase1.md`
- `raptor_progress_phase1_auth.md`
- `raptor_progress_phase1_identity.md`
- `ha-be/internal/domain/identity/user.go`
- `ha-be/internal/domain/identity/repository.go`
- `ha-be/internal/domain/identity/service.go`
- `ha-be/internal/infrastructure/persistence/user_repo.go`
- `ha-be/internal/infrastructure/persistence/user_repo_gorm.go`
- `ha-be/internal/infrastructure/email/smtp_mailer.go`
- `ha-be/internal/platform/auth/jwt.go`
- `ha-be/internal/platform/middleware/middleware.go`
- `ha-be/internal/interfaces/http/handler.go`
- `ha-be/main.go`

## Plan checklist — Phase 1

- [x] `domain/identity/` entities + interfaces — `user.go`, `repository.go` exist
- [x] `infrastructure/persistence/user_repo.go` — in-memory repo exists
- [x] `infrastructure/persistence/user_repo_gorm.go` — GORM repo exists
- [x] `internal/platform/auth/` — JWT issue/validate helpers — `jwt.go` with GenerateToken/ParseToken/RevokeToken using jwt/v5
- [ ] `application/identity/` — register, login, renew, reset, confirm use-cases — **directory does not exist**; logic baked into `domain/identity/service.go` directly
- [x] `infrastructure/email/smtp_mailer.go` — present; `net/smtp` based `SMTPMailer`; `Mailer` interface; nil-safe
- [ ] `interfaces/http/identity/` handlers — **no separate package**; all handlers in `interfaces/http/handler.go` monolith
- [x] `internal/platform/middleware/auth.go` — `Auth` middleware in `middleware.go`

## Findings vs Plan

### application/identity/ layer missing
- Plan §4 calls for `application/identity/` with use-case functions (register, login, renew, reset, confirm)
- Actual: `domain/identity/service.go` contains all business logic directly
- This violates the layering in PLAN.md §4 and §3 SOLID (Single Responsibility)
- `application/deck/` and `application/replay/` exist — inconsistency across domains

### interfaces/http/identity/ missing
- Plan §4: `interfaces/http/identity/` as dedicated package for identity handlers
- Actual: all 7+ identity handlers live in `interfaces/http/handler.go` alongside room/plugin/lfg/alert/deck/replay handlers
- Makes `handler.go` a 1170-line monolith

### ConfirmEmail uses query param, not path param
- Plan §5: `GET /be/api/v1/confirm-email/:token` (path param)
- Actual: `GET /be/api/v1/confirm-email?token=<value>` (query param)
- Minor deviation from plan route definition

### confirm/reset tokens are in-memory only
- `confirmTokens` and `resetTokens` are `map[string]tokenEntry` on `IdentityService`
- TTL expiry is implemented (`tokenEntry.ExpiresAt`) — not missing
- Tokens are lost on server restart — no persistent storage
- Progress files do not flag this as a gap

### PluginPermission handler is a stub
- `PluginPermission` in handler.go returns hardcoded `allowed: false` for GET and `status: ok` for POST/DELETE
- No actual `UserPluginPermission` domain model or persistence
- Plan §2 lists `UserPluginPermission` as part of the Plugin bounded context
- Plan §5: `GET/POST/DELETE /be/api/v1/users/plugin_permission/:plugin_id(/:user_id)` routes exist but are stubs

## Items confirmed correct (previous review had false complaints)

- [x] `infrastructure/email/smtp_mailer.go` — EXISTS; was incorrectly marked missing in previous review
- [x] `ConfirmEmail` calls `repo.Update(user)` — fixed; persistence is correct
- [x] `ResetPassword` calls `repo.Update(user)` — fixed; persistence is correct
- [x] Token TTL expiry — `tokenEntry.ExpiresAt` is set and checked
- [x] JWT library is v5 — `github.com/golang-jwt/jwt/v5 v5.3.1` in go.mod
- [x] Auth TTLs wired from config — `identitySvc.SetTokenTTL(...)` called from `main.go` using config values
- [x] `POST /be/api/v1/recaptcha/verify` — implemented; returns 501 when secret unset
- [x] Password hashing uses bcrypt (`golang.org/x/crypto/bcrypt`) — confirmed in `service.go`

## TODO

- [ ] Create `application/identity/` use-case layer wrapping domain service, per plan §4 folder structure
- [ ] Move identity HTTP handlers to `interfaces/http/identity/` package; clean up handler.go
- [ ] Fix `GET /be/api/v1/confirm-email/:token` to use path param not query param
- [ ] Implement `UserPluginPermission` domain model + persistence; replace stub handler
- [ ] Persistent token store for confirm/reset tokens (DB-backed, survive restart)

## Outcome

- [x] Review updated with accurate current implementation state
- [x] False complaints from previous review removed
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
