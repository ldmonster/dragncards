# Raptor Progress - Phase3 Offline Start

## Purpose
Track implementation status of offline-compatible behavior for the WebSocket / channel hub.

## Status — COMPLETE

### Offline POST path (`internal/interfaces/ws/ws.go`)
- [x] `WSHandler.ServeHTTP` handles `POST /be/socket` in addition to WebSocket upgrade
- [x] Decodes Phoenix envelope JSON from request body
- [x] Resolves `clientID` from `X-Client-ID` header; falls back to `offline-<timestamp>` when absent
- [x] Routes envelope through `DispatchChannelMessage` → channel handler — identical dispatch path as live WS
- [x] Encodes handler response as Phoenix envelope and writes to `http.ResponseWriter`
- [x] `client_id` injected into response payload for stateless offline ID continuity

### Supported offline events (room channel)
| Event | Behaviour |
|---|---|
| `phx_join` | Subscribe clientID to topic |
| `phx_leave` | Unsubscribe clientID from topic |
| `game_action` | Append action via `RoomService`; broadcast to topic |
| `request_state` | Return persisted action log as `send_state` |
| `reset_game` | Clear action log; broadcast `game_reset` |
| `reset_and_reload` | Clear action log; broadcast `reset_and_reload` |
| `set_seat` | Broadcast `set_seat` |
| `set_spectator` | Broadcast `set_spectator` |
| `send_alert` | Broadcast `send_alert` |
| `go_to_replay_step` | Broadcast `go_to_replay_step` |
| `step_through` | Broadcast `step_through` |

### Tests (`internal/interfaces/ws/ws_test.go`)
- [x] `TestWSHandlerPostOfflineClientIDUnique` — unique ID per request when no header
- [x] `TestWSHandlerPostOfflineGameActionBroadcastAndPersist` — action appended and broadcast via POST
- [x] `TestWSHandlerPostOfflineRequestState` — `request_state` returns persisted actions
- [x] `TestWSHandlerPostOfflineResetGameBroadcast` — `reset_game` clears state and broadcasts

### Verified items
- [x] `request_state` returns persisted actions for offline POST
- [x] Stateful events for `reset_game` path are broadcast correctly

## Implementation notes
- WS upgrade remains JWT query-token authenticated; offline POST has no auth requirement by design (headless client).
- Offline POST is stateless per request: clientID is ephemeral unless caller persists `X-Client-ID`.
- Room action state is held in `RoomRepository` (in-memory or GORM); shared between WS and offline POST paths.

## TODO
- [ ] Document `/be/socket` offline POST path and envelope schema in README
- [ ] Hook LFG and chat channel events to persistent stores where currently missing
- [ ] Add `step_through` / `go_to_replay_step` replay persistence (Phase 4 overlap)
- [x] Persist seat/spectator assignments in GameUI and compose `seats_changed`/`spectators_changed` payloads from state
