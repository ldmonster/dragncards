# Raptor Review - phase3 websocket

## Scope
Follow `review_instructions.md`: read plan, read progress files, review progress files with complaints to plan,
compare progress files with implementation. No new features — only evaluation of existing plan progress.

## Files read
- `PLAN.md` (Phase 3, §6 WebSocket Protocol)
- `raptor_progress_phase3_ws.md`
- `ha-be/internal/interfaces/ws/ws.go`
- `ha-be/internal/interfaces/ws/hub.go`
- `ha-be/internal/interfaces/ws/conn.go`
- `ha-be/internal/interfaces/ws/message.go`
- `ha-be/internal/interfaces/ws/channel_registry.go`
- `ha-be/internal/interfaces/ws/room_channel.go`
- `ha-be/internal/interfaces/ws/chat_channel.go`
- `ha-be/internal/interfaces/ws/lobby_channel.go`
- `ha-be/internal/interfaces/ws/lfg_channel.go`
- `ha-be/internal/interfaces/ws/my_topic_channel.go`
- `ha-be/internal/interfaces/http/handler.go`

## Phase-level checklist (Phase 3 WS items)

| Plan item | Status |
|---|---|
| `interfaces/ws/hub.go` - registry, topic fan-out, broadcast | ✅ implemented |
| `interfaces/ws/conn.go` - read/write pump, ping/pong | ✅ implemented; ping/pong not explicitly tested |
| `interfaces/ws/message.go` - Phoenix envelope encode/decode | ✅ implemented |
| `room_channel.go` — all plan events | ✅ all plan events handled |
| `chat_channel.go` | ✅ stub; basic broadcast |
| `lobby_channel.go` | ✅ stub; basic broadcast |
| `lfg_channel.go` | ✅ new_post/delete_post persist via LfgService |
| `my_topic_channel.go` | ✅ stub |
| WS endpoint `GET /be/socket?token=...` | ✅ JWT token auth + nhooyr.io/websocket upgrade |
| Phoenix envelope (topic/event/payload/ref/join_ref) | ✅ all fields present |
| Channel topic routing | ✅ DispatchChannelMessage routes all 5 channel types |
| `save_replay` client-to-server event | ✅ implemented in room_channel.go; persists via ReplayService |
| `users_changed` server-to-client event | ✅ `broadcastUsersChanged` called on phx_join/phx_leave |
| `seats_changed` server-to-client event | ✅ `broadcastSeatsChanged` called on set_seat |
| `spectators_changed` server-to-client event | ✅ `broadcastSpectatorsChanged` called on set_spectator |
| `gui_update` server-to-client event | ✅ `broadcastGUIUpdate` called after game_action |
| `bad_game_state` server-to-client event | ✅ returned from room_channel on game_action error |

## Complaints / gaps (as "to plan")

- WARN `channel_registry.go` is not listed in plan §4 folder structure. It is used as the dispatch entry point called from `ws.go`. It is not dead code — it is the correct place for routing logic. Not a problem, just undocumented in plan.
- WARN `chat_channel.go`, `lobby_channel.go`, `my_topic_channel.go` remain minimal stubs — no reconnect, no history, no message persistence beyond in-memory broadcast. Plan §6 channel table lists these as real channels but plan does not specify their persistence requirements.
- WARN `conn.go` write pump uses buffered channel but ping/pong heartbeat is not verified in tests.
- WARN `jwt/v4` noted as TODO in previous review — now resolved: `go.mod` uses `github.com/golang-jwt/jwt/v5 v5.3.1`.
- WARN Tests only cover offline POST path (`ws_test.go`). No tests for live WS upgrade+read+write cycle, ping/pong, or disconnect/reconnect.
- WARN `current_state` server-to-client event is listed in plan §6 channel table for room but is not generated anywhere. `request_state` returns `send_state`, not `current_state`.

## TODO

- [ ] Add WS integration tests: live upgrade, read/write, disconnect, ping/pong heartbeat.
- [ ] Add tests for chat/lobby/lfg/my_topic channels (join, message, leave).
- [ ] Clarify `current_state` vs `send_state` — plan lists both; implementation only sends `send_state`.

## Outcome

- [x] Review file updated with accurate current implementation state (re-reviewed 2026-03-20 — all findings still valid)
- [x] All previously noted TODOs resolved: jwt v5, save_replay, users_changed, seats_changed, spectators_changed, gui_update, bad_game_state — all now implemented
- [x] channel_registry.go confirmed in use — not dead code
- [x] Review points tracked with check marks
- [x] No new features suggested; findings are gaps vs the existing plan
