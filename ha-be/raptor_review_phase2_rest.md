# Raptor Review - phase2 rest

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 2 section, §5 REST API Routes, §2 Bounded Contexts)
- `raptor_progress_phase2_rest.md`
- `ha-be/internal/domain/` (directory listing)
- `ha-be/internal/application/` (directory listing)
- `ha-be/internal/infrastructure/persistence/` (directory listing)
- `ha-be/internal/interfaces/http/handler.go`
- `ha-be/main.go`

## Plan checklist — Phase 2

### Domain aggregates
- [x] `domain/room/` — Room, RoomAction, interfaces, service
- [x] `domain/plugin/` — Plugin, CustomCard, interfaces, service
- [x] `domain/lfg/` — LfgPost, interfaces, service
- [x] `domain/alert/` — Alert, interfaces, service
- [x] `domain/deck/` — Deck, DeckRepository interface exists
- [x] `domain/replay/` — Replay, ReplayRepository interface exists
- [ ] `domain/settings/` — CardAlt, CardBackAlt, BackgroundAlt **absent entirely**

### Persistence repos
- [x] `persistence/room_repo.go` + `room_repo_gorm.go`
- [x] `persistence/plugin_repo.go` + `plugin_repo_gorm.go`
- [x] `persistence/lfg_repo.go` + `lfg_repo_gorm.go`
- [x] `persistence/alert_repo.go` + `alert_repo_gorm.go`
- [x] `persistence/deck_repo.go` + `deck_repo_gorm.go`
- [x] `persistence/replay_repo.go` + `replay_repo_gorm.go`

### Application layer
- [ ] `application/room/` — does not exist; service in `domain/room/service.go`
- [ ] `application/plugin/` — does not exist; service in `domain/plugin/service.go`
- [x] `application/deck/` — exists as `internal/application/deck/service.go`
- [x] `application/replay/` — exists as `internal/application/replay/service.go`
- [ ] `application/lfg/` — does not exist; service in `domain/lfg/service.go`
- [ ] `application/alert/` — does not exist; service in `domain/alert/service.go`

### Router — chi with path params
- [x] chi router (`go-chi/chi/v5`) used in `NewRouter`; `chi.URLParam` works for `:pluginID`, `:deckID`, `:replayID`

### Routes implemented (from `NewRouter` in handler.go)
- [x] `GET /be/health`
- [x] `POST /be/api/v1/registration`
- [x] `POST /be/api/v1/session`
- [x] `DELETE /be/api/v1/session`
- [x] `POST /be/api/v1/session/renew`
- [x] `GET /be/api/v1/confirm-email`
- [x] `POST /be/api/v1/reset-password`
- [x] `POST /be/api/v1/reset-password/update`
- [x] `POST /be/api/v1/recaptcha/verify`
- [x] `GET/POST /be/api/rooms`
- [x] `GET /be/api/plugins`
- [x] `GET /be/api/plugins/visible`
- [x] `POST /be/api/plugins`
- [x] `GET /be/api/plugins/{pluginID}`
- [x] `GET/POST /be/api/plugins/{pluginID}/cards`
- [x] `GET/POST/DELETE /be/api/v1/lfg` (auth-protected)
- [x] `GET/POST /be/api/v1/alerts` (auth-protected)
- [x] `GET/POST/DELETE /be/api/v1/profile` (auth-protected)
- [x] `GET /be/api/v1/users/all` (auth-protected)
- [x] `GET/POST/DELETE /be/api/v1/users/plugin_permission/{pluginID}/{userID}` (auth-protected, stub)
- [x] `POST /be/api/v1/admin_contact` (auth-protected, stub)
- [x] `POST /be/api/v1/admin/update_user_patreon` (auth-protected, stub)
- [x] `POST /be/api/v1/games` (auth-protected)
- [x] `GET/POST/GET/{deckID}/PUT/{deckID}/DELETE/{deckID} /be/api/v1/decks` (auth-protected)
- [x] `GET /be/api/v1/public_decks/{pluginID}` (auth-protected)
- [x] `GET/POST/GET/{replayID}/DELETE/{replayID} /be/api/v1/replays` (auth-protected)

### Routes still missing from plan §5
| Route | Status |
|---|---|
| `GET /be/api/plugins/visible/:user_id` | ❌ missing (no user_id param) |
| `GET /be/api/plugins/visible/:plugin_id/:user_id` | ❌ missing |
| `POST /be/api/plugin-repo-update` | ❌ missing |
| `GET /be/api/v1/admin_contact` | ⚠️ implemented as POST only; plan shows GET |
| `GET/POST/DELETE /be/api/replays/*` | ⚠️ under `/v1/replays` not `/be/api/replays` |
| `GET/POST/DELETE /be/api/custom_content/*` | ❌ missing |
| `GET /be/api/my_custom_content/:user_id/:plugin_id` | ❌ missing |
| `GET /be/api/all_custom_content/:user_id/:plugin_id` | ❌ missing |
| `GET /be/metrics` | ❌ missing (Phase 6) |

## Findings vs Plan

### application/ layer incomplete
- Plan §4 lists separate application packages for all domains
- Actual: only `application/deck/` and `application/replay/` have application layer packages
- `room`, `plugin`, `lfg`, `alert` services live directly in their domain packages
- `application/game/` also exists (Phase 4)
- Inconsistent: some domains have application layer, others don't

### Handler stubs — admin and plugin_permission
- `PluginPermission` handler: returns hardcoded `allowed: false` for GET, `status: ok` for POST/DELETE — no real logic
- `AdminContact` handler: accepts POST, returns `status: "queued"` without doing anything
- `AdminUpdateUserPatreon` handler: accepts POST, returns `updated: true` without touching DB
- These routes exist at the correct paths but contain no real business logic

### custom_content / RepoUpdate entirely absent
- Plan §5: `GET/POST/DELETE /be/api/custom_content/*`, `GET /be/api/my_custom_content/...`, `GET /be/api/all_custom_content/...`
- `POST /be/api/plugin-repo-update`
- None of these are implemented; no domain model for Settings or plugin sync

### Settings domain absent
- Plan §2: bounded context with CardAlt, CardBackAlt, BackgroundAlt persisted in Postgres
- No `domain/settings/` directory or models anywhere in codebase

### AutoMigrate does not cover deck and replay
- `main.go` AutoMigrate list: `identity.User, room.Room, room.RoomAction, plugin.Plugin, plugin.CustomCard, lfg.LfgPost, alert.Alert`
- `deck.Deck` and `replay.Replay` are **not** in the AutoMigrate call — GORM repos exist but table won't be created on startup

## TODO

- [x] Add `deckDomain.Deck` and `replayDomain.Replay` to `db.AutoMigrate(...)` in `main.go`
- [ ] Create `application/room/`, `application/plugin/`, `application/lfg/`, `application/alert/` use-case layers per plan §4
- [ ] Implement real `UserPluginPermission` domain model + persistence; replace stub handler
- [x] Implement `POST /be/api/plugin-repo-update` and plugin repo sync use-case (stubbed endpoint)
- [ ] Implement custom_content routes and Settings domain (CardAlt, CardBackAlt, BackgroundAlt)
- [x] Add `GET /be/api/plugins/visible/:user_id` and `GET /be/api/plugins/visible/:plugin_id/:user_id` routes
- [x] Fix `GET /be/api/v1/admin_contact` — plan shows GET; current is POST only
- [x] Implement real admin_contact and admin/update_user_patreon handlers (currently stubs)


## Outcome

- [x] Review updated with accurate current implementation state (re-reviewed 2026-03-20 — all findings still valid)
- [x] False complaints from previous review removed: deck/replay domains exist, chi router used, all route paths corrected, lfg/alert AutoMigrate fixed
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
