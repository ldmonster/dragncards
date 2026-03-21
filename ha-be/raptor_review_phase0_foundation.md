# Raptor Review - phase0 foundation

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 0 section, §4, §9, §10, §11, §12, §13)
- `raptor_progress_phase0_foundation.md`
- `ha-be/go.mod`
- `ha-be/main.go`
- `ha-be/config/config.go`
- `ha-be/internal/platform/logger/logger.go`
- `ha-be/internal/platform/database/database.go`
- `ha-be/internal/platform/redis/redis.go`
- `ha-be/internal/platform/middleware/middleware.go`
- `ha-be/cmd/root.go`, `cmd/serve.go`, `cmd/migrate.go`

## Plan checklist — Phase 0

- [x] `go mod init` in ha-be/ — `go.mod` exists with correct module path
- [x] `config/config.go` — real YAML parser using `gopkg.in/yaml.v3`; full struct (server, database, redis, auth, email, patreon, recaptcha, log)
- [x] `internal/platform/logger/` — custom `Logger` struct with `Trace()` (Level -8) and `Fatal()` (Level 12); wraps `*slog.Logger`
- [x] `internal/platform/database/` — GORM + PostgreSQL, `WithTx()` and `RawQuery[T]()` both present
- [x] `internal/platform/redis/` — real `go-redis/v9` client; pings on init; nil-safe when URL empty
- [x] `internal/platform/middleware/` — `RequestID` (header passthrough + generated ID), `Logger` (duration log), `Recovery` (panic → 500), `Auth` (Bearer JWT)
- [x] `internal/interfaces/http/router.go` — chi router (`go-chi/chi/v5`) with CORS, RequestID, Logger, Recovery middleware chain in `NewRouter`
- [x] `GET /be/health` endpoint — returns `{"status":"ok"}`
- [x] `main.go` wiring — config load, DB connect/migrate, service init, chi router, WS handler, graceful shutdown
- [x] cobra CLI (`cmd/root.go`, `cmd/serve.go`, `cmd/migrate.go`) — structure present, root.go wires AddCommand
- [x] `go-chi/chi/v5` in go.mod ✅
- [x] `gopkg.in/yaml.v3` in go.mod ✅
- [x] `github.com/redis/go-redis/v9` in go.mod ✅
- [x] `github.com/spf13/cobra` in go.mod ✅
- [x] `github.com/golang-jwt/jwt/v5` in go.mod ✅ (v5, matches plan §12)
- [x] `github.com/golang-migrate/migrate/v4` in go.mod — `v4.17.1` ✅; `migrations/` directory with SQL files added
- [x] `cmd/serve.go` fully implemented — calls `RunServe()` with config flag
- [x] `cmd/migrate.go` fully implemented — runs `migrate.Up()` with proper file source
- [x] cobra `cmd.Execute()` called from `main.go` — `main.go` is now just `cmd.Execute()`
- [x] `golang-migrate` SQL migration files — `migrations/000001_create_schema.up.sql` and down file added; CLI applies


## Findings vs Plan

### cobra CLI — fully active
- `main.go` delegates entirely to `cmd.Execute()` (single import + call; no flags)
- `cmd/serve.go` calls `RunServe(serveCfgPath)` from `cmd/serve_impl.go`; `--config` flag wired
- `cmd/migrate.go` calls `migrate.New("file://"+migrationSourcePath, cfg.Database.URL).Up()` — real `golang-migrate` logic; `--config` and `--path` flags wired
- `cmd/root.go` adds both commands via `init()`; `Execute()` exported and used from `main.go`
- All progress-file claims confirmed correct

### golang-migrate present; SQL migration files exist
- `github.com/golang-migrate/migrate/v4 v4.17.1` in `go.mod`
- `migrations/000001_create_schema.up.sql` and `migrations/000001_create_schema.down.sql` present
- `cmd/migrate.go` applies migrations via `golang-migrate` CLI path
- `db.AutoMigrate(...)` still called in `cmd/serve_impl.go` on startup as an additional safety net — acceptable duplication; does not break plan requirement

### go.mod dependency status vs plan §12
| Planned | In go.mod | Status |
|---|---|---|
| `github.com/go-chi/chi/v5` | ✅ v5.2.5 | ok |
| `gorm.io/gorm` | ✅ v1.31.1 | ok |
| `gorm.io/driver/postgres` | ✅ v1.6.0 | ok |
| `nhooyr.io/websocket` | ✅ v1.8.17 | ok |
| `github.com/golang-jwt/jwt/v5` | ✅ v5.3.1 | ok |
| `github.com/golang-migrate/migrate/v4` | ✅ v4.17.1 | ok |
| `github.com/spf13/cobra` | ✅ v1.10.2 | ok |
| `github.com/redis/go-redis/v9` | ✅ v9.18.0 | ok |
| `go.opentelemetry.io/otel` | ❌ absent | missing (Phase 6) |
| `github.com/prometheus/client_golang` | ❌ absent | missing (Phase 6) |
| `golang.org/x/crypto` | ✅ v0.45.0 | ok |
| `gopkg.in/yaml.v3` | ✅ v3.0.1 | ok |

Note: OTel and Prometheus are Phase 6 items — not expected in Phase 0.

## TODO

(nothing outstanding — all plan Phase 0 items confirmed implemented)

## Outcome

- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-21 — cobra active, golang-migrate present, all findings resolved)
- [x] Previously reported gaps resolved: cobra CLI now fully active (`cmd.Execute()` called from `main.go`); `cmd/serve.go` and `cmd/migrate.go` both fully implemented; `golang-migrate v4.17.1` in `go.mod`; SQL migration files in `migrations/`
- [x] Previously reported false gaps corrected: config YAML, logger Trace/Fatal, database WithTx/RawQuery, redis real client, chi router, CORS, middleware, jwt v5 — all implemented
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are deviations from existing plan
