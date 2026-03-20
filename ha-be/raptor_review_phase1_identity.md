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
- [ ] `application/identity/` use-case layer — **does not exist**; all logic stays in `domain/identity/service.go`
- [ ] `interfaces/http/identity/` package — **does not exist**; all handlers in monolithic `handler.go`
- [ ] `GET /be/api/v1/confirm-email/:token` — implemented as query param `?token=` not path segment (minor deviation from plan)

## Findings vs Plan

### application/identity/ layer still missing
- Plan §4: `application/identity/` with use-case functions wrapping domain service
- Actual: `domain/identity/service.go` contains all business logic directly
- Violates plan layering (§3 SOLID, §4 Folder Structure)
- Neither progress file acknowledges this gap

### interfaces/http/identity/ still missing
- Plan: `interfaces/http/identity/` as a dedicated package
- Actual: all identity handlers remain in `interfaces/http/handler.go` (1170 lines, all domains mixed)
- Makes handler.go increasingly monolithic

### confirm-email route: query param vs path param
- Plan: `GET /be/api/v1/confirm-email/:token`
- Actual: `GET /be/api/v1/confirm-email?token=<value>`
- Minor deviation; functionally equivalent but does not match plan spec

### Token revocation is still in-memory only
- Both `confirmTokens` and `resetTokens` on `IdentityService` are in-memory maps
- Lost on server restart; no DB or Redis persistence
- Plan §7 notes stateless approach for JWT, but confirm/reset tokens are not JWTs — they are string tokens stored in service memory

## TODO

- [ ] Create `application/identity/` with use-case functions per plan folder structure
- [ ] Move identity HTTP handlers to `interfaces/http/identity/` package
- [ ] Change confirm-email route to path param `:token` to match plan §5
- [ ] Persist confirm/reset tokens to DB (store on `users` table or separate table) to survive restarts

## Outcome

- [x] Review file updated with accurate current implementation state
- [x] Previously reported false gaps corrected: email mailer, ConfirmEmail/ResetPassword persistence, token TTL, JWT v5, recaptcha endpoint, bcrypt, auth TTL from config — all now implemented
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
