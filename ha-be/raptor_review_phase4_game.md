# Raptor Review - phase4 game

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features - only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 4, §8, §3, §4)
- `raptor_progress_phase4_game.md`
- `ha-be/main.go` (re-read 2026-03-20)
- `ha-be/internal/domain/game/state.go`
- `ha-be/internal/domain/game/card.go`
- `ha-be/internal/domain/game/stack.go`
- `ha-be/internal/domain/game/group.go`
- `ha-be/internal/domain/game/player_info.go`
- `ha-be/internal/domain/game/evaluate/evaluator.go`
- `ha-be/internal/domain/game/evaluate/context.go`
- `ha-be/internal/domain/game/evaluate/function.go`
- `ha-be/internal/domain/game/evaluate/variable.go`
- `ha-be/internal/domain/game/evaluate/functions/registry.go`
- `ha-be/internal/domain/game/evaluate/functions/arithmetic.go`
- `ha-be/internal/domain/game/evaluate/functions/noop.go`
- `ha-be/internal/application/game/game_service.go`
- `ha-be/internal/application/game/room_registry.go`
- `ha-be/internal/infrastructure/gamestate/store.go`
- `ha-be/internal/infrastructure/gamestate/redis_store.go`
- `ha-be/internal/infrastructure/gamestate/postgres_store.go`
- `ha-be/main.go`

## Plan checklist — Phase 4

| Plan item | Status |
|---|---|
| `domain/game/state.go` — GameUI aggregate root | [x] exists; has Slug, Players, Actions, Cards, Stacks, Groups, Infos |
| `domain/game/card.go` | [x] exists |
| `domain/game/stack.go` | [x] exists |
| `domain/game/group.go` | [x] exists |
| `domain/game/player_info.go` | [x] exists |
| `domain/game/evaluate/evaluator.go` — dispatcher | [x] implemented |
| `domain/game/evaluate/context.go` — EvalContext | [x] exists; typed accessors for Card/Stack/Group/PlayerInfo |
| `domain/game/evaluate/function.go` — GameFunction interface | [x] `Execute(ctx *EvalContext, args []any)` — type-safe |
| `domain/game/evaluate/variable.go` — GameVariable interface | [x] standalone file; `Resolve(ctx *EvalContext)` — type-safe |
| `domain/game/evaluate/functions/` — ~100 DSL ops | [~] only `noop`, `add`, `sub`, `mul`, `div` (5 of ~100) |
| `application/game/room_registry.go` | [x] in-memory map[slug]*GameRoom with `Close()` / `Done()` |
| `application/game/game_service.go` | [x] CreateGame, SendAction, GetGameUI, ResetGame, run goroutine with `ctx.Done()` and `Done()` cancel paths |
| `infrastructure/gamestate/store.go` — GameStateStore interface | [x] exists: Save, Load, Delete |
| `infrastructure/gamestate/redis_store.go` | [x] exists; stores GameUI JSON in Redis under `gamestate:<slug>` |
| `infrastructure/gamestate/postgres_store.go` | [x] exists; stores GameUI as JSONB in `game_states` table |
| WS to GameService wiring | [x] `room_channel.go` calls `gameSvc.SendAction`; `gameSvc` passed in `main.go` |
| AppendAction errors logged | [x] `s.log.Error(...)` — not silently dropped |
| GameRoom.Close() + Done() for clean shutdown | [x] present in `room_registry.go` |

## Findings vs Plan

- WARN ~95 of the ~100 planned DSL functions are not yet implemented. Only `noop` + 4 arithmetic builtins are registered. The evaluator framework is solid and open/closed for addition, but the actual game logic cannot run without the remaining functions. Progress file tracks this as current in-progress TODO.
- OK `infrastructure/gamestate/` is now fully wired in `main.go`: if `cfg.Redis.URL` is set, a Redis client is initialised and `NewRedisGameStateStore` is used; otherwise `NewPostgresGameStateStore` is used; in-memory fallback when neither DB nor Redis is available. Fixed as of 2026-03-20.
- OK `game_states` table is created by `pgStore.AutoMigrate()` called in `main.go` (line 107) whenever a Postgres store is selected. Fixed as of 2026-03-20.

## TODO

- [ ] Port remaining ~95 DSL functions into `evaluate/functions/` per plan §8.
- [x] Add store selection logic in `main.go` — Redis/Postgres/in-memory selection implemented (2026-03-20).
- [x] `game_states` table created via `pgStore.AutoMigrate()` in `main.go` (2026-03-20).
- [x] Redis client initialised in `main.go` when `cfg.Redis.URL` is non-empty (2026-03-20).

## Outcome
- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-20).
- [x] Previously reported false gaps resolved: card/stack/group/player_info models exist; GameUI has all maps; EvalContext has typed accessors; variable.go is standalone; GameFunction and GameVariable use *EvalContext; gamestate stores all exist; ctx.Done() and Close() present; AppendAction errors logged.
- [x] Stale TODOs resolved: stateStore selection logic wired; game_states AutoMigrate wired; Redis client init wired — all three fixed in main.go.
- [x] Review points tracked with check marks.
- [x] No new features suggested; findings are gaps vs the existing plan.
