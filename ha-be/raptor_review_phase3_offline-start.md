# Raptor Review - phase3 offline-start

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features - only evaluation of existing plan progress.

## Files read
- `PLAN.md`
- `raptor_progress_phase3_offline-start.md`
- `raptor_progress_phase3_ws.md`
- `ha-be/internal/interfaces/ws/ws.go`
- `ha-be/internal/interfaces/ws/ws_test.go`
- `ha-be/internal/interfaces/ws/room_channel.go`
- `ha-be/internal/application/game/game_service.go`
- `ha-be/main.go`

## Plan checklist — Phase 3

- [x] Offline POST path at `POST /be/socket` — implemented in `ws.go`
- [x] `X-Client-ID` header for offline client ID; falls back to `offline-<timestamp>`
- [x] `DispatchChannelMessage` dispatch path shared between WS and offline POST
- [x] `phx_join` / `phx_leave` — topic subscribe/unsubscribe
- [x] `game_action` — persisted via `gameSvc.SendAction`; broadcast `send_update` to topic
- [x] `request_state` — returns action log as `send_state` from `gameSvc.GetGameUI` or fallback `svc.ListActions`
- [x] `reset_game` / `reset_and_reload` — clears state via `gameSvc.ResetGame`; broadcast to topic
- [x] `set_seat` — broadcast + `seats_changed` event generated
- [x] `set_spectator` — broadcast + `spectators_changed` event generated
- [x] `send_alert`, `go_to_replay_step`, `step_through` — broadcast to topic
- [x] `save_replay` — persisted via `replaySvc.Save` in `room_channel.go`
- [x] `users_changed` generated on phx_join / phx_leave via `broadcastUsersChanged`
- [x] `gui_update` generated on game_action via `broadcastGUIUpdate`
- [x] `bad_game_state` returned when `game_action` fails
- [x] `AppendAction` errors logged via `s.log.Error(...)` — not silently dropped
- [x] `gameSvc` passed to `NewWSHandler` in `main.go` — end-to-end wiring complete
- [x] Tests: `TestWSHandlerPostOfflineClientIDUnique`, `TestWSHandlerPostOfflineGameActionBroadcastAndPersist`, `TestWSHandlerPostOfflineRequestState`, `TestWSHandlerPostOfflineResetGameBroadcast`

## Findings vs Plan

- WARN `PLAN.md` has no mention of an offline POST fallback path. The offline-start feature is a branch-specific addition not tracked in PLAN.md at all. Progress files document it but plan never acknowledges it.
- WARN `set_seat` and `set_spectator` broadcast the raw payload but do not mutate application-level seat/spectator state in `GameUI` or any domain model. The generated `seats_changed` / `spectators_changed` events carry subscriber IDs only, not actual seat assignments.
- WARN `chat`, `lobby`, `lfg`, `my_topic` channels are minimal stubs — basic broadcast only for offline POST, no message history, no reconnect handling.
- WARN Offline POST `game_action` requires a game room to already exist in the registry (created via WS `phx_join` or `POST /be/api/v1/games`); the error message "room not found" is returned but omits any guidance on how to create the room.

## TODO

- [~] Add offline-start feature note to `PLAN.md` — progress tracker row mentions `+ offline POST path` but §13 Phase 3 implementation checklist does not document it at all; minimal mention only.
- [x] Persist seat/spectator assignments in `GameUI` when `set_seat` / `set_spectator` are received so `seats_changed` / `spectators_changed` carry actual state.
- [ ] Add tests for chat/lobby/lfg/my_topic offline POST flows and WS reconnect handling.
- [ ] Clarify `game_action` error response when game room does not exist yet.

## Outcome
- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-21 — all open findings still valid).
- [x] ``PLAN.md`` offline POST minimally mentioned in progress tracker but not in §13 Phase 3 checklist.
- [x] Review points tracked with check marks.
- [x] No new features suggested; findings are gaps vs the existing plan.
