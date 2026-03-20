# Raptor Progress - Phase4 Game Engine

## Purpose
Track core game engine DSL evaluation and room game loop implementation.

## Status — IN PROGRESS

### Domain model (`internal/domain/game/state.go`)
- [x] `GameUI` struct with player/action tracking fields

### Eval framework (`internal/domain/game/evaluate/`)
- [x] `Evaluator` — `RegisterFunction`, `RegisterVariable`, `EvalFunction`, `GetVariable`
- [x] `EvalContext` — execution context passed to all functions and variable resolvers
- [x] `GameFunction` interface — `Name() string`, `Execute(ctx, args) (any, error)`
- [x] `GameVariable` interface — `Name() string`, `Resolve(ctx) (any, error)`
- [x] `NoopFunction` — reference no-op implementation
- [x] `evaluate/functions/contract.go` — `asInt` helper
- [x] `evaluate/functions/registry.go` — `RegisterDefaults(e *Evaluator)` helper registers builtins
- [x] Arithmetic builtins registered: `add`, `sub`, `mul`, `div`
- [x] Unit tests passing (`internal/domain/game/evaluate/evaluator_test.go`)

### Application layer (`internal/application/game/`)
- [x] `RoomRegistry` — in-memory map of room slug → `*GameRoom`
- [x] `GameService` — `CreateGame(roomSlug)`, `SendAction(roomSlug, action)` with action persistence via `room.RoomService.AppendAction`
- [x] `GameService` tests (`game_service_test.go`) — create game, send action, persist action

### Persistence
- [x] `room.RoomAction` GORM model; `RoomRepository.AppendAction` + `ListActions`
- [x] GORM impl in `persistence/room_repo_gorm.go`; in-memory impl in `persistence/room_repo.go`

## Implementation notes
- `GameService` uses `room.RoomService` for action persistence — no separate game-state table yet.
- DSL function set is minimal (4 arithmetic ops); full ~100-function port from Elixir backend is pending.
- `RoomRegistry` is in-memory; game state is lost on server restart — DB/Redis backing is planned.

## TODO
- [~] Port full DSL function set (~100 operations) from Elixir evaluator — core expression evaluators now implemented in `evaluate/engine.go` (LIST, AND/OR/NOT, EQUAL/NOT_EQUAL/GT/GTE/LT/LTE, arithmetic, IN_STRING, JOIN_STRING, OBJ_GET_BY_PATH fallback, MAP, FILTER, ONE_CARD, FOR_EACH_KEY_VAL, VAR, PREV, COND, WHILE, MOVE_CARD)
- [x] Implement variable definitions and variable resolver registry
- [ ] DB-backed game state store (Redis when REDIS_URL present; Postgres JSON otherwise)
- [ ] Wire `GameService` into `WSHandler` room channel so `game_action` events drive the evaluator
- [ ] Goroutine-per-room game loop with action replay on reconnect
- [ ] Phase 5: plugin extensibility and custom card DB integration
