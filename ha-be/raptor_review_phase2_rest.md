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
- [x] `domain/settings/` — `internal/domain/settings/settings.go` + `persistence/settings_repo.go` + `persistence/settings_repo_gorm.go` exist; `Setting` model covers card_alt, card_back_alt, background_alt

### Persistence repos
- [x] `persistence/room_repo.go` + `room_repo_gorm.go`
- [x] `persistence/plugin_repo.go` + `plugin_repo_gorm.go`
- [x] `persistence/lfg_repo.go` + `lfg_repo_gorm.go`
- [x] `persistence/alert_repo.go` + `alert_repo_gorm.go`
- [x] `persistence/deck_repo.go` + `deck_repo_gorm.go`
- [x] `persistence/replay_repo.go` + `replay_repo_gorm.go`

### Application layer
- [ ] `application/room/` — does not exist; service in `domain/room/service.go`
- [x] `application/plugin/` — `internal/application/plugin/service.go` exists
- [x] `application/deck/` — `internal/application/deck/service.go` exists
- [x] `application/replay/` — `internal/application/replay/service.go` exists
- [ ] `application/lfg/` — does not exist; service in `domain/lfg/service.go`
- [ ] `application/alert/` — does not exist; service in `domain/alert/service.go`
- [x] `application/settings/` — `internal/application/settings/service.go` exists

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
| `GET /be/api/plugins/visible/:user_id` | ✅ `GET /be/api/plugins/visible/{userID}` implemented |
| `GET /be/api/plugins/visible/:plugin_id/:user_id` | ✅ `GET /be/api/plugins/visible/{pluginID}/{userID}` implemented |
| `POST /be/api/plugin-repo-update` | ✅ implemented (stub: returns 202 queued; no real sync) |
| `GET /be/api/v1/admin_contact` | ✅ GET and POST both implemented |
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

### Handler stubs — remaining
- `PluginRepoUpdate` handler: accepts POST, returns `{"status": "plugin repo update request enqueued"}` (202) — no actual repo sync logic
- `AdminContact` GET: returns `{"status": "ready", "message": "admin contact endpoint"}` — informational stub
- `AdminUpdateUserPatreon` handler: parses `user_id` + `tier` from body; returns `{"updated": true}` — no DB persistence
- `PluginPermission` handlers now call real `pluginSvc.GetUserPluginPermission`, `SetUserPluginPermission`, `DeleteUserPluginPermission` — not stubs

### custom_content entirely absent; plugin repo update is a stub
- Plan §5: `GET/POST/DELETE /be/api/custom_content/*`, `GET /be/api/my_custom_content/...`, `GET /be/api/all_custom_content/...` — none implemented
- `POST /be/api/plugin-repo-update` — route exists and returns HTTP 202; handler body has no actual sync logic (fetches nothing, stores nothing)
- No domain model for custom content

### Settings domain — now present
- `internal/domain/settings/settings.go`: `Setting` model with `UserID`, `PluginID`, `CardAlt`, `CardBackAlt`, `BackgroundAlt` GORM fields
- `internal/application/settings/service.go` + `persistence/settings_repo.go` + `persistence/settings_repo_gorm.go` all exist
- Settings endpoints available: `GET/POST/DELETE /be/api/v1/settings/{userID}/{pluginID}`

### AutoMigrate — comprehensive coverage
- `cmd/serve_impl.go` AutoMigrate list: `identity.User, room.Room, room.RoomAction, plugin.Plugin, plugin.CustomCard, plugin.UserPluginPermission, lfg.LfgPost, alert.Alert, deck.Deck, replay.Replay, settings.Setting`
- All domain models covered including deck, replay, settings, and UserPluginPermission

## TODO

- [x] Add `deckDomain.Deck` and `replayDomain.Replay` to `db.AutoMigrate(...)` — now in `cmd/serve_impl.go`
- [ ] Create `application/room/`, `application/lfg/`, `application/alert/` use-case layers per plan §4 (`application/plugin/` and `application/settings/` now done)
- [x] Implement real `UserPluginPermission` domain model + persistence; replace stub handler
- [ ] Implement real plugin repo sync in `PluginRepoUpdate` (currently 202 stub)
- [x] Implement settings domain and endpoints (CardAlt, CardBackAlt, BackgroundAlt)
- [x] Add `GET /be/api/plugins/visible/{userID}` and `GET /be/api/plugins/visible/{pluginID}/{userID}` routes
- [x] Fix `GET /be/api/v1/admin_contact` — GET and POST both implemented
- [ ] Implement real `AdminUpdateUserPatreon` with DB persistence (currently parses input but returns stub `updated: true`)
- [ ] Implement `GET/POST/DELETE /be/api/custom_content/*`, `GET /be/api/my_custom_content/:user_id/:plugin_id`, `GET /be/api/all_custom_content/:user_id/:plugin_id`

## Outcome

- [x] Review updated with accurate current implementation state (re-reviewed 2026-03-21)
- [x] Resolved since last review: settings domain, application/plugin, visible/user_id routes, admin_contact GET, plugin_permission real handlers, AutoMigrate full coverage
- [x] Still open: application/room, application/lfg, application/alert; custom_content routes; AdminUpdateUserPatreon stub; PluginRepoUpdate stub
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
