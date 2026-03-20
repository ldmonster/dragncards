# Raptor Review - phase3 offline-start

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md`
- `raptor_progress_phase3_offline-start.md`
- `raptor_progress_phase3_ws.md`
- `ha-be/internal/interfaces/ws/ws.go`
- `ha-be/internal/interfaces/ws/ws_test.go`
- `ha-be/internal/interfaces/ws/room_channel.go`
- `ha-be/main.go`

## Plan checklist — Phase 3

- [x] Phase 0 Foundation markers in `PLAN.md` — marked `[x]`
- [x] Phase 1 Auth markers in `PLAN.md` — marked `[x]`
- [x] Phase 2 REST API markers in `PLAN.md` — marked `[x]`
- [x] Phase 3 WS markers in `PLAN.md` — marked `[x]`

## Offline POST path — implementation status

| Item | Status |
|---|---|
| `POST /be/socket` accepted alongside WS `GET /be/socket` | ✅ |
| Phoenix envelope JSON decoded from request body | ✅ |
| `X-Client-ID` header for offline client ID; `offline-<timestamp>` fallback | ✅ |
| Routes through `DispatchChannelMessage` — same path as WS | ✅ |
| Response encoded as Phoenix envelope | ✅ |
| `client_id` injected into response payload | ✅ |
| `game_action` requires game room to exist (via WS phx_join or POST /be/api/v1/games) | ✅ (by design) |

## Supported offline events (room channel)

| Event | Behaviour |
|---|---|
| `phx_join` | Subscribe clientID to topic |
| `phx_leave` | Unsubscribe clientID from topic |
| `game_action` | Persist via GameService + broadcast |
| `request_state` | Return action log as `send_state` |
| `reset_game` | GameService.ResetGame + broadcast |
| `reset_and_reload` | Broadcast |
| `set_seat` | Broadcast + seats_changed |
| `set_spectator` | Broadcast + spectators_changed |
| `send_alert` | Broadcast |
| `go_to_replay_step` | Broadcast |
| `step_through` | Broadcast |
| `save_replay` | Persist via ReplayService + broadcast |

## Tests (`ws_test.go`)

- [x] `TestWSHandlerPostOfflineClientIDUnique`
- [x] `TestWSHandlerPostOfflineGameActionBroadcastAndPersist`
- [x] `TestWSHandlerPostOfflineRequestState`
- [x] `TestWSHandlerPostOfflineResetGameBroadcast`

## Complaints / gaps vs plan

- WARN The offline POST path is not mentioned anywhere in `PLAN.md`. It is a branch-specific addition (`feat/offline-start`). Progress files document it but the plan never acknowledges it. This is not a code bug, but the plan is out of sync with what was built.
- WARN `game_action` on offline POST: if the game room does not yet exist, a `bad_game_state` reply is returned with the error message. The error is surfaced but the error text is not particularly descriptive for the offline caller. No game room auto-creation path from POST.

## TODO

- [ ] Document the offline POST path in `PLAN.md` under Phase 3 or a dedicated section so the plan reflects the branch feature
- [ ] Add tests for chat/lobby/lfg/my_topic offline POST event flows

## Outcome

- [x] Review updated with accurate current implementation state
- [x] All previous TODO items verified resolved: save_replay persists, AppendAction errors surfaced via log, all room events handled, GameService wired, tests present
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are gaps vs the existing plan
