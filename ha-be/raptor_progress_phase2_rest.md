# Raptor Progress - Phase2 REST

## Purpose
Track REST domain endpoint implementation progress for Phase2.

## Status — COMPLETE (core domains)

### Domains implemented

#### Room (`internal/domain/room/`)
- [x] `Room` + `RoomAction` GORM models
- [x] `RoomRepository` interface with `Create`, `List`, `FindBySlug`, `AppendAction`, `ListActions`
- [x] In-memory impl (`persistence/room_repo.go`) + GORM impl (`persistence/room_repo_gorm.go`)
- [x] `RoomService` — `Create`, `List`, `AppendAction`
- [x] Endpoints: `GET /be/api/rooms`, `POST /be/api/rooms`

#### Plugin (`internal/domain/plugin/`)
- [x] `Plugin` + `CustomCard` + `UserPluginPermission` GORM models
- [x] `PluginRepository` interface with `Create`, `List`, `ListVisible`, `FindByID`, `CreateCustomCard`, `ListCustomCards`, `CreatePermission`, `GetPermission`, `DeletePermission`
- [x] In-memory + GORM impls
- [x] `PluginService` — `Create`, `List`, `ListVisible`, `CreateCustomCard`, `ListCustomCards`, `GetUserPluginPermission`, `SetUserPluginPermission`, `DeleteUserPluginPermission`
- [x] Endpoints: `GET /be/api/plugins`, `GET /be/api/plugins/visible`, `POST /be/api/plugins`, `GET/POST /be/api/plugins/{pluginID}/cards`, `GET/POST/DELETE /be/api/v1/users/plugin_permission/{pluginID}/{userID}`
- [x] Unit tests passing (`main_test.go` extension)

#### LFG (`internal/domain/lfg/`)
- [x] `LfgPost` GORM model
- [x] `LfgRepository` interface with `Create`, `List`, `Delete`
- [x] In-memory + GORM impls
- [x] `LfgService` — `Create`, `List`, `Delete`
- [x] Endpoints (auth-protected): `GET /be/api/lfg/`, `POST /be/api/lfg/`, `DELETE /be/api/lfg/`

#### Alert (`internal/domain/alert/`)
- [x] `Alert` GORM model
- [x] `AlertRepository` interface with `Create`, `List`
- [x] In-memory + GORM impls
- [x] `AlertService` — `Create`, `List`
- [x] Endpoints (auth-protected): `GET /be/api/alerts/`, `POST /be/api/alerts/`

### Router (`NewRouter` in `internal/interfaces/http/handler.go`)
- [x] chi router with global CORS, RequestID, Logger, Recovery middleware

### Settings (`internal/domain/settings/` + `internal/application/settings/`)
- [x] `Setting` model with `card_alt`, `card_back_alt`, `background_alt`
- [x] Settings repository interface + in-memory + GORM impls
- [x] Settings service for get/upsert/list/delete
- [x] Endpoints: `GET /be/api/v1/settings/{userID}/{pluginID}`, `POST /be/api/v1/settings`, `DELETE /be/api/v1/settings/{userID}/{pluginID}`
- [x] `settings` exposed in legacy routes `/be/api/settings/...`
- [x] Auth middleware applied to lfg + alerts via `r.Group`
- [x] chi URL params used for `{pluginID}` sub-routes

### Persistence
- [x] All five domains have both in-memory and GORM-backed implementations
- [x] `main.go` selects GORM (PostgreSQL) when DB connects; falls back to in-memory
- [x] `db.AutoMigrate` covers `identity.User`, `room.Room`, `room.RoomAction`, `plugin.Plugin`, `plugin.CustomCard`, `lfg.LfgPost`, `alert.Alert`, `deck.Deck`, `replay.Replay`

### Health
- [x] `GET /be/health` — returns `{"status":"ok"}`

## TODO
- [x] Add `lfg.LfgPost` and `alert.Alert` to `AutoMigrate` in `main.go`
- [x] Admin / replay / profile / deck / settings endpoints (Phase 5+)
- [ ] Pagination and input validation middleware
- [ ] Structured error JSON responses
