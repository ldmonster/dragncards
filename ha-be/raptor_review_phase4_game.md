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
| `domain/game/evaluate/functions/` — ~100 DSL ops | [~] 79 of ~100 functions registered; files: arithmetic, comparison, collection, collection_advanced, collection_basic, collection_group, advanced, advanced_control, string_ops, concat, plugin, noop |
| `application/game/room_registry.go` | [x] in-memory map[slug]*GameRoom with `Close()` / `Done()` |
| `application/game/game_service.go` | [x] CreateGame, SendAction, GetGameUI, ResetGame, run goroutine with `ctx.Done()` and `Done()` cancel paths |
| `infrastructure/gamestate/store.go` — GameStateStore interface | [x] exists: Save, Load, Delete |
| `infrastructure/gamestate/redis_store.go` | [x] exists; stores GameUI JSON in Redis under `gamestate:<slug>` |
| `infrastructure/gamestate/postgres_store.go` | [x] exists; stores GameUI as JSONB in `game_states` table |
| WS to GameService wiring | [x] `room_channel.go` calls `gameSvc.SendAction`; `gameSvc` passed in `main.go` |
| AppendAction errors logged | [x] `s.log.Error(...)` — not silently dropped |
| GameRoom.Close() + Done() for clean shutdown | [x] present in `room_registry.go` |

## Files read (added 2026-03-21)
- `ha-be/internal/domain/game/evaluate/engine.go`
- `ha-be/internal/domain/game/evaluate/tests/engine_arithmetic_test.go`
- `ha-be/internal/domain/game/evaluate/tests/engine_collection_test.go`
- `ha-be/internal/domain/game/evaluate/tests/engine_dsl_test.go`
- `ha-be/internal/domain/game/evaluate/functions/` (all 16 files counted)

## Findings vs Plan

- OK full DSL function set substantially implemented: 79 functions registered across 16 source files covering arithmetic, comparison, boolean logic, string ops, collection ops (map/filter/reduce), group ops, object path read/write, control flow (cond/while), game ops (move_card, one_card, for_each_key_val), var/prev, and plugin ops.
- OK `engine.go` exists as the full expression evaluation entry point; handles raw list literals, nested command fallback, root command dispatch.
- OK expanded test suite: `evaluate/tests/engine_arithmetic_test.go`, `engine_collection_test.go`, `engine_dsl_test.go` — 3 test files, all passing.
- OK `infrastructure/gamestate/` is now fully wired in `cmd/serve_impl.go`: if `cfg.Redis.URL` is set, a Redis client is initialised and `NewRedisGameStateStore` is used; otherwise `NewPostgresGameStateStore` is used; in-memory fallback when neither DB nor Redis is available.
- OK `game_states` table is created by `pgStore.AutoMigrate()` called in `cmd/serve_impl.go` whenever a Postgres store is selected.

## TODO

- [ ] Port remaining ~21 DSL functions into `evaluate/functions/` per plan §8 (79 of ~100 registered as of 2026-03-21).
- [x] Add store selection logic in `cmd/serve_impl.go` — Redis/Postgres/in-memory selection implemented.
- [x] `game_states` table created via `pgStore.AutoMigrate()` in `cmd/serve_impl.go`.
- [x] Redis client initialised in `cmd/serve_impl.go` when `cfg.Redis.URL` is non-empty.
- [x] Confirmed fix for replay on reconnect in `GameService.CreateGame`.
- [ ] Phase 5 plugin extension work still pending.

## Outcome
- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-21).
- [x] DSL function count updated from 5 to 79; engine.go and expanded test suite confirmed.
- [x] All store selection and AutoMigrate wiring confirmed in `cmd/serve_impl.go` (previously in `main.go`).
- [x] Review points tracked with check marks.
- [x] No new features suggested; findings are gaps vs the existing plan.
