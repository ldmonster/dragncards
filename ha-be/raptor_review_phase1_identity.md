# Raptor Review - phase1 identity

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 1 section, §2 Identity context, §7 Auth Design)
- `raptor_progress_phase1.md`, `raptor_progress_phase1_auth.md`, `raptor_progress_phase1_identity.md`
- `ha-be/internal/domain/identity/user.go`, `repository.go`, `service.go`
- `ha-be/internal/infrastructure/persistence/user_repo.go`, `user_repo_gorm.go`
- `ha-be/internal/infrastructure/email/smtp_mailer.go`
- `ha-be/internal/platform/auth/jwt.go`
- `ha-be/internal/platform/middleware/middleware.go`
- `ha-be/internal/interfaces/http/handler.go`
- `ha-be/main.go`

## Plan checklist — Phase 1

- [x] `domain/identity/` entities + interfaces — `user.go`, `repository.go` exist
- [x] `infrastructure/persistence/user_repo.go` — in-memory repo present
- [x] `infrastructure/persistence/user_repo_gorm.go` — GORM repo present
- [x] `internal/platform/auth/` — JWT issue/validate/revoke; `jwt/v5` ✅
- [x] `infrastructure/email/smtp_mailer.go` — `net/smtp` based `SMTPMailer` with `Mailer` interface; nil-safe when unconfigured
- [x] `internal/platform/middleware/auth.go` — `Auth` middleware in `middleware.go`; injects `user_id` into context
- [x] `ConfirmEmail` calls `repo.Update(user)` after setting `user.Confirmed = true` ✅
- [x] `ResetPassword` calls `repo.Update(user)` after setting new password hash ✅
- [x] Token TTL via `tokenEntry{UserID, ExpiresAt}` — confirm and reset tokens both expire ✅
- [x] Auth TTL wired from config — `main.go` reads `cfg.Auth.AccessLifetime` / `cfg.Auth.RefreshLifetime` and passes to handler and identity service ✅
- [x] `POST /be/api/v1/recaptcha/verify` — `VerifyRecaptcha` handler implemented ✅
- [x] Password hashing uses bcrypt (`golang.org/x/crypto/bcrypt`) ✅
- [x] `application/identity/` use-case layer — `internal/application/identity/service.go` exists; wraps domain service
- [ ] `interfaces/http/identity/` package — **does not exist**; all handlers in monolithic `handler.go`
- [x] `GET /be/api/v1/confirm-email/:token` — `GET /be/api/v1/confirm-email/{token}` implemented as path param; legacy query param route also present

## Findings vs Plan

### Progress files incorrectly document SHA-256 hashing
- `raptor_progress_phase1_identity.md` says: "Password hashing (SHA-256 + salt; golang.org/x/crypto for constant-time compare)"
- `raptor_progress_phase1_auth.md` says: "Password hashing uses SHA-256 + random salt (not bcrypt) to avoid CGO in offline builds"
- Actual: `internal/domain/identity/service.go` imports and uses `golang.org/x/crypto/bcrypt`; `bcrypt.GenerateFromPassword` and `bcrypt.CompareHashAndPassword` are called — not SHA-256
- Progress files are wrong; implementation uses bcrypt (plan §1 Auth Design: "bcrypt via golang.org/x/crypto/bcrypt" ✅)

### application/identity/ layer still missing
- Plan §4: `application/identity/` with use-case functions wrapping domain service
- Actual: `domain/identity/service.go` contains all business logic directly
- Still a plan-form architecture gap (SOLID layering) but not urgent for behavior correctness

### interfaces/http/identity/ still missing
- Plan: `interfaces/http/identity/` as a dedicated package
- Actual: all identity handlers remain in `interfaces/http/handler.go` (1170 lines, all domains mixed)
- Can be refactored later; this follow-up does not address it

### confirm-email route: path param now implemented
- Plan: `GET /be/api/v1/confirm-email/:token`
- Updated to support `/be/api/v1/confirm-email/{token}` with legacy query fallback

### token persistence now implemented
- Confirm/reset tokens moved from in-memory maps to `users` table columns (`confirm_token`, `confirm_token_expires_at`, `reset_token`, `reset_token_expires_at`)
- `IdentityService` now uses repository lookups by token and clears token fields after use

### token persistence — now DB-backed
- Confirm/reset tokens are columns on `users` table: `ConfirmToken`, `ConfirmTokenExpiresAt`, `ResetToken`, `ResetTokenExpiresAt` (GORM `varchar(256)` + `index`)
- `IdentityService` uses repository lookups by token and clears token fields after use
- No in-memory maps remain; tokens survive server restart

## TODO

- [x] Create `application/identity/` with use-case functions per plan folder structure (`internal/application/identity/service.go` exists)
- [ ] Move identity HTTP handlers to `interfaces/http/identity/` package (all handlers still in monolithic `handler.go`)
- [x] Change confirm-email route to path param `:token` to match plan §5 (already implemented, with legacy query fallback)
- [x] Persist confirm/reset tokens to DB — `ConfirmToken`, `ConfirmTokenExpiresAt`, `ResetToken`, `ResetTokenExpiresAt` are GORM columns on `User`; no in-memory maps remain

## Outcome

- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-21)
- [x] `application/identity/service.go` confirmed present — TODO resolved
- [x] Token persistence confirmed: confirm/reset tokens stored in DB columns, no in-memory maps — TODO resolved
- [x] Still open: `interfaces/http/identity/` package; all identity handlers remain in monolithic `handler.go`
- [x] Progress file SHA-256 vs bcrypt discrepancy still valid — implementation correctly uses bcrypt; progress files remain inaccurate
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
