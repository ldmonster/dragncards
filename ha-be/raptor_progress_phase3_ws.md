# Raptor Progress - Phase3 WebSocket

## Purpose
Track real-time WebSocket channel implementation for Phase3.

## Status — COMPLETE (core)

### Files (`internal/interfaces/ws/`)
- `ws.go` — `WSHandler` serving both GET (WS upgrade) and POST (offline envelope) at `/be/socket`
- `hub.go` — `Hub` with `Register`, `Unregister`, `BroadcastToTopic`, topic map (nil-safe)
- `conn.go` — `Conn` wrapping `nhooyr.io/websocket`; buffered `send` channel; `runWritePump` goroutine
- `message.go` — `PhoenixEnvelope` (topic/event/payload/ref); `Marshal`/`Unmarshal`
- `channel_registry.go` — `DispatchChannelMessage` routing to per-topic channel handlers
- `room_channel.go` — handles `phx_join`, `phx_leave`, `game_action`, `request_state`, `reset_game`, `reset_and_reload`, `set_seat`, `set_spectator`, `send_alert`, `go_to_replay_step`, `step_through`
- `chat_channel.go` — handles `chat:*` topic messages
- `lobby_channel.go` — handles `lobby` topic messages
- `lfg_channel.go` — handles `lfg` topic messages; `new_post`/`delete_post` persist via `LfgService`
- `my_topic_channel.go` — handles `my_topic:*` per-user messages
- `ws_test.go` — unit tests for hub broadcast, offline POST path, room/action flows

### WebSocket upgrade (GET `/be/socket?token=...`)
- [x] JWT query-token auth via `auth.ParseToken`
- [x] `nhooyr.io/websocket` upgrade with `InsecureSkipVerify` CORS
- [x] Per-connection goroutine with `runWritePump` and context-cancelled read loop
- [x] Unique `clientID`: `userID` or `anon-<timestamp>`; suffix timestamp if ID already registered (multi-tab support)
- [x] Phoenix envelope JSON (`{topic, event, payload, ref}`) read/write loop

### Offline POST path (POST `/be/socket`)
- [x] Accepts Phoenix envelope JSON body
- [x] `X-Client-ID` header or time-based fallback for offline client ID
- [x] Routes through `DispatchChannelMessage` → channel handlers exactly like WS path
- [x] Returns Phoenix envelope response (e.g. `request_state` → `send_state`)
- [x] `client_id` included in response payload for offline ID tracking

### Events handled in room channel
- [x] `phx_join` / `phx_leave` — topic subscribe/unsubscribe
- [x] `game_action` — appended via `room.RoomService.AppendAction`; broadcast to topic; replay queue support started
- [x] `request_state` — returns current action log as `send_state`
- [x] `reset_game` / `reset_and_reload` — clears action log; broadcast
- [x] `set_seat` / `set_spectator` — broadcast to topic
- [x] `send_alert` — broadcast to topic
- [x] `go_to_replay_step` / `step_through` — broadcast to topic
- [x] `save_replay` — persisted through `replay.ReplayService.Save` when replay payload is valid

### Tests (`ws_test.go`)
- [x] `TestWSHandlerPostOfflineClientIDUnique`
- [x] `TestWSHandlerPostOfflineGameActionBroadcastAndPersist`
- [x] `TestWSHandlerPostOfflineRequestState`
- [x] `TestWSHandlerPostOfflineResetGameBroadcast`

## TODO
- [ ] Integration tests for chat/lobby/lfg WS events and reconnect handling
- [x] Connect room channel events to `GameService` room loop (Phase 4 overlap)
- [ ] Admin replay enqueue on `game_action` events
