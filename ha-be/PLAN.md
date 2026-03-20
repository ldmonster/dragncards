
# DragnCards Go Backend — Architecture & Implementation Plan

> **File purpose**: track design decisions and implementation progress.
> Update status markers as work proceeds.

Original logic - backend folder (written on elixir)

---

## Status Legend
- `[ ]` Not started
- `[~]` In progress
- `[x]` Done

---

## Progress Tracker

```
[x] Phase 0  - Foundation (module scaffold, config, logger, DB, CLI)
[x] Phase 1  - Auth domain (users, sessions, tokens, email)
[x] Phase 2  - REST API domains (rooms, plugins, LFG, alerts) — core done; decks/replays/admin pending
[x] Phase 3  - WebSocket / game hub (real-time room, chat, lobby, lfg channels + offline POST path)
[x] Phase 4  - Game engine DSL interpreter (eval framework + full core builtins from Elixir + error behavior + tests)
[ ] Phase 5  - Plugin system (custom card DB, plugin repo sync)
[ ] Phase 6  - Observability (OTel traces, Prometheus metrics)
[ ] Phase 7  - Tests and final wire-up
```

---

## 1. Confirmed Stack Decisions

| Concern | Choice | Source |
|---|---|---|
| HTTP middleware | chi pipeline only; all handlers raw net/http | required |
| ORM / DB | gorm + PostgreSQL; gorm.DB.Raw() for custom SQL | required |
| Logging | log/slog + custom Fatal() + Trace level | required |
| CLI | cobra | required |
| Config | YAML file, manually parsed - no viper | required |
| WebSocket library | nhooyr.io/websocket (actively maintained, context-native) | Q1b |
| WS wire protocol | Phoenix channel JSON envelope compatible - frontend JS untouched | Q1 |
| Auth tokens | golang-jwt/jwt - stateless JWT pair; logout client-side only | Q2+Q6 |
| Email | Plain SMTP via net/smtp | Q3 |
| Game state backing | Per-room goroutine; Redis if REDIS_URL env set, else Postgres JSON save | Q4 |
| DB migrations | golang-migrate (SQL files) | Q5 |
| Observability | slog + OpenTelemetry traces + Prometheus /metrics | Q7 |

---

## 2. Bounded Contexts (DDD)

| Context | Aggregates | Persistence |
|---|---|---|
| Identity | User, TokenPair | Postgres users |
| Room | Room, RoomLog | Postgres rooms, room_logs |
| Plugin | Plugin, UserPluginPermission, CustomCardDb | Postgres |
| Deck | Deck | Postgres decks |
| Replay | Replay | Postgres replays |
| Game (runtime) | GameUI, Card, Stack, Group, PlayerInfo | Goroutine + Redis/Postgres backing |
| Chat | ChatMessage | In-memory ring buffer per room |
| LFG | LfgPost, LfgResponse, LfgSubscription | Postgres lfg_* |
| Alerts | Alert | Postgres alerts |
| Settings | CardAlt, CardBackAlt, BackgroundAlt | Postgres |

---

## 3. SOLID Principles

| Principle | Application |
|---|---|
| S - Single Responsibility | Each handler does one thing; business logic in application layer; data access in repos |
| O - Open/Closed | GameFunction interface - add DSL ops without touching evaluator |
| L - Liskov | All repos implement their port interface; swappable with mocks in tests |
| I - Interface Segregation | Separate Reader/Writer/Querier interfaces per aggregate |
| D - Dependency Inversion | Handlers depend on service interfaces injected in main.go |

---

## 4. Folder Structure

```
ha-be/
|-- cmd/
|   |-- root.go
|   |-- serve.go
|   `-- migrate.go
|-- config/
|   |-- config.go           # YAML -> Config struct (no viper)
|   `-- config.yaml.example
|-- internal/
|   |-- platform/
|   |   |-- logger/         # slog + Fatal + Trace level
|   |   |-- database/       # gorm init, WithTx(), RawQuery[T]()
|   |   |-- middleware/     # chi: Auth, Logger, Recovery, CORS, RequestID
|   |   |-- auth/           # JWT issue/validate helpers
|   |   `-- redis/          # optional Redis client init
|   |-- domain/
|   |   |-- identity/
|   |   |   |-- user.go     # User entity
|   |   |   |-- token.go    # TokenPair value object
|   |   |   |-- repository.go  # UserRepository interface
|   |   |   `-- service.go  # IdentityService interface
|   |   |-- room/           # Room, RoomLog, interfaces
|   |   |-- plugin/         # Plugin, Permission, CustomCardDb, interfaces
|   |   |-- deck/
|   |   |-- replay/
|   |   |-- lfg/            # LfgPost, LfgResponse, LfgSubscription, interfaces
|   |   |-- alert/
|   |   `-- game/
|   |       |-- state.go    # GameUI aggregate root
|   |       |-- card.go
|   |       |-- stack.go
|   |       |-- group.go
|   |       |-- player_info.go
|   |       `-- evaluate/
|   |           |-- evaluator.go   # dispatcher
|   |           |-- context.go     # EvalContext
|   |           |-- function.go    # GameFunction interface
|   |           |-- variable.go    # GameVariable interface
|   |           `-- functions/     # ~100 files, one per DSL op
|   |-- application/
|   |   |-- identity/       # register, login, renew_token, reset_password, confirm_email
|   |   |-- room/
|   |   |-- plugin/
|   |   |-- deck/
|   |   |-- replay/
|   |   |-- lfg/
|   |   |-- alert/
|   |   `-- game/
|   |       |-- room_registry.go  # in-memory map[slug]*GameRoom
|   |       `-- game_service.go  # CreateRoom, GameAction, StepThrough, ...
|   |-- infrastructure/
|   |   |-- persistence/    # GORM repos for all aggregates
|   |   |-- email/          # smtp_mailer.go
|   |   `-- gamestate/
|   |       |-- redis_store.go
|   |       `-- postgres_store.go
|   `-- interfaces/
|       |-- http/
|       |   |-- router.go   # chi mux + middleware chain + route registration
|       |   |-- respond/    # JSON helpers, error mapping
|       |   |-- identity/   # register, session, profile, password handlers
|       |   |-- room/
|       |   |-- plugin/
|       |   |-- deck/
|       |   |-- replay/
|       |   |-- lfg/
|       |   |-- alert/
|       |   `-- admin/
|       `-- ws/
|           |-- hub.go             # connection registry, broadcast, topic routing
|           |-- conn.go            # per-connection read/write pump
|           |-- message.go         # Phoenix envelope struct
|           |-- room_channel.go
|           |-- chat_channel.go
|           |-- lobby_channel.go
|           |-- lfg_channel.go
|           `-- my_topic_channel.go
`-- main.go
```

---

## 5. REST API Routes (prefix: /be)

### Public
```
POST   /be/api/v1/registration
POST   /be/api/v1/session
DELETE /be/api/v1/session
POST   /be/api/v1/session/renew
GET    /be/api/v1/confirm-email/:token
POST   /be/api/v1/reset-password
POST   /be/api/v1/reset-password/update
POST   /be/api/v1/recaptcha/verify
GET    /be/api/rooms
POST   /be/api/rooms
GET    /be/api/plugins
GET    /be/api/plugins/:plugin_id
GET    /be/api/plugins/visible/:user_id
GET    /be/api/plugins/visible/:plugin_id/:user_id
POST   /be/api/plugin-repo-update
GET    /be/health
GET    /be/metrics
```

### Protected (JWT required)
```
GET/POST/DELETE /be/api/v1/profile/*
GET             /be/api/v1/users/all
GET/POST/DELETE /be/api/v1/users/plugin_permission/:plugin_id(/:user_id)
POST            /be/api/v1/games
GET/POST/PUT/DELETE /be/api/v1/decks/*
GET             /be/api/v1/public_decks/:plugin_id
GET             /be/api/v1/alerts
GET             /be/api/v1/admin_contact
POST            /be/api/v1/admin/update_user_patreon
GET/POST/DELETE /be/api/v1/lfg/*
GET/POST/DELETE /be/api/replays/*
GET/POST/PUT/DELETE /be/api/custom_content/*
GET             /be/api/my_custom_content/:user_id/:plugin_id
GET             /be/api/all_custom_content/:user_id/:plugin_id
```

---

## 6. WebSocket Protocol

Endpoint: GET /be/socket (HTTP upgrade).
Auth: JWT in query param ?token=<auth_token>.

Phoenix-compatible envelope:
```json
{ "topic": "room:abc", "event": "game_action", "payload": {}, "ref": "1", "join_ref": "1" }
```

### Channels
| Topic | Client -> Server | Server -> Client |
|---|---|---|
| room:{slug} | phx_join, game_action, save_replay, step_through, set_seat, set_spectator, reset_game, reset_and_reload, request_state, send_alert | current_state, send_update, send_state, gui_update, users_changed, seats_changed, spectators_changed, go_to_replay_step, send_alert, bad_game_state |
| chat:{slug} | phx_join, new_msg | new_msg |
| lobby:lobby | phx_join, leave_room | lobby_update |
| lfg:{plugin_id} | phx_join, new_post, delete_post | new_post, delete_post |
| my_topic:{user_id} | phx_join | new_message |

---

## 7. Auth Design

- Token pair: auth_token (30m TTL) + renew_token (90d TTL), both signed JWTs
- Revocation: stateless - logout is client-side only
- Password: bcrypt via golang.org/x/crypto/bcrypt
- Renew flow: expired auth_token + valid renew_token -> new pair

---

## 8. Game Engine Architecture

GameFunction interface (Open/Closed):
```go
type GameFunction interface {
    Name() string
    Execute(ctx *EvalContext, args []any) (any, error)
}
```

Evaluator registers all ~100 DSL functions at startup.
Each lives in domain/game/evaluate/functions/<name>.go.

GameRoom goroutine:
```go
type GameRoom struct {
    Slug  string
    State *domain.GameUI
    In    chan RoomMessage
}
func (r *GameRoom) Run(ctx context.Context) {
    for { select { case <-ctx.Done(): return; case msg := <-r.In: r.dispatch(msg) } }
}
```

State store (pluggable via REDIS_URL env):
```go
type GameStateStore interface {
    Save(ctx context.Context, slug string, state *domain.GameUI) error
    Load(ctx context.Context, slug string) (*domain.GameUI, error)
}
```

---

## 8.1. Follow-up plan
- Completed: all core game DSL ops from Elixir now implemented and tested (arithmetic, comparators, boolean, list, map/filter, object path read/write, reduce, rand, set, one_card, for_each_key_val, var, prev, cond, while, move_card).
- Remaining: evaluate full 100+ op set (incl. card-specific game actions, random draw semantics, penalty logic, event queue).
- Next: implement plugin integration (Phase 5) with card DB lookup and custom rules evaluation via same DSL engine.
- Next: wire WS `game_action` pipeline to evaluator in room goroutine, confirm no race conditions in action replay.
- Next: add formal error mode in evaluator for JSON schema and invalid DSL command reporting (for UI debugging).
- Next: complete persistence path: Redis store in Redis mode + periodic Postgres to disk; add snapshots.

---

## 9. Logger
```go
const LevelTrace = slog.Level(-8)
const LevelFatal = slog.Level(12)

func (l *Logger) Trace(msg string, args ...any) { l.inner.Log(ctx, LevelTrace, msg, args...) }
func (l *Logger) Fatal(msg string, args ...any) { l.inner.Log(ctx, LevelFatal, msg, args...); os.Exit(1) }
```

---

## 10. GORM Helpers
```go
func WithTx(db *gorm.DB, fn func(tx *gorm.DB) error) error {
    tx := db.Begin()
    defer func() { if r := recover(); r != nil { tx.Rollback() } }()
    if err := fn(tx); err != nil { tx.Rollback(); return err }
    return tx.Commit().Error
}

func RawQuery[T any](db *gorm.DB, sql string, args ...any) ([]T, error) {
    var out []T
    return out, db.Raw(sql, args...).Scan(&out).Error
}
```

---

## 11. Config YAML
```yaml
server:
  host: "0.0.0.0"
  port: 4000
database:
  host: "localhost"
  port: 5432
  name: "dragncards_dev"
  user: "postgres"
  password: "postgres"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
redis:
  url: ""   # empty = disabled; use Postgres fallback for game state
auth:
  secret: "change-me"
  auth_token_ttl: "30m"
  renew_token_ttl: "2160h"
email:
  from: "noreply@dragncards.com"
  smtp_host: "smtp.example.com"
  smtp_port: 587
  smtp_user: ""
  smtp_password: ""
log:
  level: "info"
  format: "json"
patreon:
  client_id: ""
  client_secret: ""
recaptcha:
  secret: ""
```

---

## 12. go.mod Dependencies (planned)

```
github.com/go-chi/chi/v5
gorm.io/gorm
gorm.io/driver/postgres
nhooyr.io/websocket
github.com/golang-jwt/jwt/v5
github.com/golang-migrate/migrate/v4
github.com/spf13/cobra
github.com/redis/go-redis/v9           (optional, only if Redis configured)
go.opentelemetry.io/otel
go.opentelemetry.io/otel/exporters/prometheus
github.com/prometheus/client_golang
golang.org/x/crypto
gopkg.in/yaml.v3
```

---

## 13. Implementation Phases

### Phase 0 - Foundation
- [ ] go mod init in ha-be/
- [ ] cmd/root.go, cmd/serve.go, cmd/migrate.go (cobra)
- [ ] config/config.go - YAML parser
- [ ] internal/platform/logger/ - slog + Fatal + Trace
- [ ] internal/platform/database/ - gorm + WithTx + RawQuery
- [ ] internal/platform/redis/ - optional client
- [ ] internal/interfaces/http/router.go - chi + middleware
- [ ] GET /be/health endpoint
- [ ] main.go wiring

### Phase 1 - Identity
- [ ] domain/identity/ entities + interfaces
- [ ] infrastructure/persistence/user_repo.go
- [ ] internal/platform/auth/ - JWT issue/validate
- [ ] application/identity/ - register, login, renew, reset, confirm
- [ ] infrastructure/email/smtp_mailer.go
- [ ] interfaces/http/identity/ handlers
- [ ] internal/platform/middleware/auth.go

### Phase 2 - REST Domains
For each: room, plugin, deck, replay, lfg, alert, admin, custom_content:
- [ ] domain/<context>/ entities + interfaces
- [ ] infrastructure/persistence/<context>_repo.go
- [ ] application/<context>/ use-cases
- [ ] interfaces/http/<context>/ handlers
- [ ] Wire routes in router.go

### Phase 3 - WebSocket Hub
- [ ] interfaces/ws/hub.go (registry, topic fan-out, broadcast)
- [ ] interfaces/ws/conn.go (read/write pump, ping/pong)
- [ ] interfaces/ws/message.go (Phoenix envelope encode/decode)
- [ ] room_channel.go, chat_channel.go, lobby_channel.go, lfg_channel.go, my_topic_channel.go

### Phase 4 - Game Engine
- [ ] domain/game/ - state, card, stack, group, player types
- [ ] domain/game/evaluate/ - evaluator + context + interfaces
- [ ] ~100 DSL functions in domain/game/evaluate/functions/
- [ ] application/game/ - room_registry + game_service
- [ ] infrastructure/gamestate/ - redis_store + postgres_store

### Phase 5 - Plugin System
- [ ] Plugin + CustomCardDb + UserPluginPermission repos
- [ ] Plugin repo sync (HTTP fetch from repo_url, store to DB)
- [ ] cmd/update_plugin.go cobra command

### Phase 6 - Observability
- [ ] OTel SDK init (OTLP exporter or stdout for dev)
- [ ] Prometheus exporter + GET /be/metrics
- [ ] Instrument HTTP handlers (request duration, status codes)
- [ ] Instrument WS hub (active connections, messages/sec)
- [ ] Instrument game engine (actions/sec, active rooms)

### Phase 7 - Tests
- [ ] Unit: DSL evaluator functions
- [ ] Unit: application services (mock repos via interfaces)
- [ ] Integration: HTTP handlers (net/http/httptest)
- [ ] Integration: WS channels (nhooyr.io/websocket test client)
- [ ] Load test: game_action throughput

---

## 14. Not Ported (by design)

| Elixir feature | Reason |
|---|---|
| OTP Application/Supervisor/GenServer | Go goroutines + channels |
| Mnesia | Replaced by stateless JWT + optional Redis |
| Pow auth library | Reimplemented as JWT |
| luerl (Lua runtime) | Not used in current plugin DSL; skip unless needed |
| Phoenix HTML views | Frontend stays React; Go serves JSON only |
| Mix tasks | Replaced by cobra CLI commands |

---

*Last updated: All decisions confirmed (Q1-Q7). Phase 0 not yet started.*

