# Raptor Progress - Phase0 Foundation

## Status — COMPLETE
- [x] Module scaffold for `ha-be` (go.mod)
- [x] `main.go` — full server wiring: config load, DB connect/migrate, service init, router mount, graceful shutdown via `os/signal` + context cancel
- [x] Configuration (`config/config.go`, `config/config.yaml.example`) — YAML parsed with `gopkg.in/yaml.v3`, struct-based with fallback defaults
- [x] Logging wrapper (`internal/platform/logger/logger.go`) — `slog`-backed with custom Trace/Fatal levels
- [x] Platform middleware (`internal/platform/middleware/middleware.go`) — RequestID, Logger, Recovery, Auth (Bearer JWT) with user-ID context injection
- [x] Platform auth (`internal/platform/auth/jwt.go`) — HS256 JWT via `golang-jwt/jwt v4`; `SetSecret`, `GenerateToken`, `ParseToken`, `RevokeToken`; in-memory revocation store (`sync.RWMutex`)
- [x] Platform database (`internal/platform/database/database.go`) — GORM + PostgreSQL, `WithTx`, generic `RawQuery[T]`
- [x] Platform redis (`internal/platform/redis/redis.go`) — optional `go-redis/v9` client; safe no-op when unconfigured
- [x] chi router (`go-chi/chi v5`) with global CORS and middleware chain in `NewRouter`
- [x] Cobra CLI (`cmd/root.go`, `cmd/serve.go`, `cmd/migrate.go`)
- [x] `db.AutoMigrate` in `main.go` migrates all GORM models on startup
- [x] All tests green — `go test ./...` passes (2026-03-20)

## Dependencies (go.mod)
| Package | Version | Role |
|---|---|---|
| `golang-jwt/jwt/v4` | v4.5.2 | JWT tokens |
| `gorm.io/driver/postgres` | v1.6.0 | PostgreSQL driver |
| `gorm.io/gorm` | v1.31.1 | ORM |
| `nhooyr.io/websocket` | v1.8.17 | WebSocket |
| `go-chi/chi/v5` | v5.2.5 | HTTP router |
| `go-chi/cors` | v1.2.2 | CORS |
| `golang-migrate/migrate/v4` | v4.19.1 | SQL migrations (cmd) |
| `redis/go-redis/v9` | v9.18.0 | Redis client |
| `spf13/cobra` | v1.10.2 | CLI |
| `golang.org/x/crypto` | v0.45.0 | Password hashing |
| `gopkg.in/yaml.v3` | v3.0.1 | Config YAML |

## Notes
- DB connection is optional: server falls back to in-memory stores when `database.URL` is unset or unreachable.
- JWT secret resolves: `cfg.Auth.JWTSecret` (YAML) → `AUTH_TOKEN_SECRET` env → hardcoded dev fallback.
- All phases build and test without an active database or Redis.
