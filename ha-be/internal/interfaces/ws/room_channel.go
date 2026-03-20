package ws

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	replayDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
)

func IsRoomTopic(topic string) bool {
	return strings.HasPrefix(topic, "room:")
}

// HandleRoomChannel dispatches all room:* channel events.
// senderConn is the originating client connection (nil on offline POST path).
func HandleRoomChannel(h *Hub, svc *room.RoomService, gameSvc *game.GameService, replaySvc *replay.ReplayService, clientID string, env PhoenixEnvelope, senderConn *Conn) (*PhoenixEnvelope, error) {
	roomSlug := strings.TrimPrefix(env.Topic, "room:")

	// Helper: write reply inline to ensure it arrives before any broadcasts.
	// Returns true when the reply was delivered directly (caller should return nil).
	sendReply := func(resp *PhoenixEnvelope) bool {
		if senderConn == nil || resp == nil {
			return false
		}
		if b, err := resp.Marshal(); err == nil {
			_ = senderConn.Send(b)
		}
		return true
	}

	switch env.Event {
	case "phx_join":
		h.Subscribe(env.Topic, clientID, h.clients[clientID])
		reply := &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}
		if sendReply(reply) {
			// send current state only to the joining peer (on live WS path)
			if statePayload, stateErr := buildRoomStatePayload(roomSlug, gameSvc, svc); stateErr == nil {
				_ = sendReply(&PhoenixEnvelope{Topic: env.Topic, Event: "current_state", Payload: statePayload})
			}
			broadcastUsersChanged(h, env.Topic)
			return nil, nil
		}
		broadcastUsersChanged(h, env.Topic)
		return reply, nil

	case "phx_leave":
		h.Unsubscribe(env.Topic, clientID)
		reply := &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}
		if sendReply(reply) {
			broadcastUsersChanged(h, env.Topic)
			return nil, nil
		}
		broadcastUsersChanged(h, env.Topic)
		return reply, nil

	case "game_action":
		var actionErr error
		if gameSvc != nil {
			actionErr = gameSvc.SendAction(roomSlug, env.Payload)
		} else if svc != nil {
			actionErr = svc.AppendAction(roomSlug, env.Payload)
		}
		if actionErr != nil {
			badStatePayload, _ := json.Marshal(map[string]string{"error": actionErr.Error()})
			return &PhoenixEnvelope{Topic: env.Topic, Event: "bad_game_state", Payload: badStatePayload, Ref: env.Ref}, nil
		}
		update := PhoenixEnvelope{Topic: env.Topic, Event: "send_update", Payload: env.Payload}
		if msg, err := update.Marshal(); err == nil {
			h.BroadcastToTopic(env.Topic, msg)
		}
		broadcastGUIUpdate(h, gameSvc, env.Topic, roomSlug)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	case "request_state":
		payload, err := buildRoomStatePayload(roomSlug, gameSvc, svc)
		if err != nil {
			return nil, err
		}
		return &PhoenixEnvelope{Topic: env.Topic, Event: "current_state", Payload: payload}, nil

	case "reset_game", "reset_and_reload":
		if gameSvc != nil {
			_ = gameSvc.ResetGame(context.Background(), roomSlug)
		}
		forward, _ := PhoenixEnvelope{Topic: env.Topic, Event: env.Event, Payload: env.Payload}.Marshal()
		h.BroadcastToTopic(env.Topic, forward)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	case "set_seat":
		var req struct {
			PlayerID string `json:"player_id"`
			Seat     string `json:"seat"`
		}
		_ = json.Unmarshal(env.Payload, &req)
		if gameSvc != nil && req.PlayerID != "" {
			_ = gameSvc.SetSeat(roomSlug, req.PlayerID, req.Seat)
		}
		forward, _ := PhoenixEnvelope{Topic: env.Topic, Event: env.Event, Payload: env.Payload}.Marshal()
		h.BroadcastToTopic(env.Topic, forward)
		broadcastSeatsChanged(h, gameSvc, env.Topic, roomSlug)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	case "set_spectator":
		var req struct {
			PlayerID  string `json:"player_id"`
			Spectator bool   `json:"spectator"`
		}
		_ = json.Unmarshal(env.Payload, &req)
		if gameSvc != nil && req.PlayerID != "" {
			_ = gameSvc.SetSpectator(roomSlug, req.PlayerID, req.Spectator)
		}
		forward, _ := PhoenixEnvelope{Topic: env.Topic, Event: env.Event, Payload: env.Payload}.Marshal()
		h.BroadcastToTopic(env.Topic, forward)
		broadcastSpectatorsChanged(h, gameSvc, env.Topic, roomSlug)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	case "send_alert", "go_to_replay_step", "step_through":
		forward, _ := PhoenixEnvelope{Topic: env.Topic, Event: env.Event, Payload: env.Payload}.Marshal()
		h.BroadcastToTopic(env.Topic, forward)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	case "save_replay":
		// Attempt to save replay metadata and events.
		if replaySvc != nil {
			var req struct {
				ID     string          `json:"id"`
				Meta   json.RawMessage `json:"meta"`
				Data   json.RawMessage `json:"data"`
				Public bool            `json:"public"`
			}
			if err := json.Unmarshal(env.Payload, &req); err == nil && req.ID != "" {
				ownerID := clientID
				if ownerID == "" {
					ownerID = "anon"
				}
				if replayObj, err := replayDomain.NewReplay(req.ID, roomSlug, ownerID, req.Meta, req.Data, req.Public); err == nil {
					_ = replaySvc.Save(replayObj)
				}
			}
		}
		forward, _ := PhoenixEnvelope{Topic: env.Topic, Event: env.Event, Payload: env.Payload}.Marshal()
		h.BroadcastToTopic(env.Topic, forward)
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil

	default:
		return nil, errors.New("unsupported room event")
	}
}

func buildRoomStatePayload(roomSlug string, gameSvc *game.GameService, svc *room.RoomService) ([]byte, error) {
	if gameSvc != nil {
		if gameUI, err := gameSvc.GetGameUI(roomSlug); err == nil && gameUI != nil {
			state := struct {
				Actions [][]byte `json:"actions"`
			}{Actions: gameUI.Actions}
			return json.Marshal(state)
		} else if err != nil {
			return nil, err
		}
	}
	if svc == nil {
		return json.RawMessage(`{"actions":[]}`), nil
	}
	actions, err := svc.ListActions(roomSlug)
	if err != nil {
		return nil, err
	}
	state := struct {
		Actions [][]byte `json:"actions"`
	}{Actions: actions}
	return json.Marshal(state)
}

func broadcastUsersChanged(h *Hub, topic string) {
	users := h.TopicSubscribers(topic)
	payload, _ := json.Marshal(map[string]any{"users": users})
	msg, _ := PhoenixEnvelope{Topic: topic, Event: "users_changed", Payload: payload}.Marshal()
	h.BroadcastToTopic(topic, msg)
}

func broadcastSeatsChanged(h *Hub, gameSvc *game.GameService, topic, roomSlug string) {
	var payload []byte
	if gameSvc != nil {
		if gameUI, err := gameSvc.GetGameUI(roomSlug); err == nil && gameUI != nil {
			payload, _ = json.Marshal(map[string]any{"seats": gameUI.Seats})
		}
	}
	if payload == nil {
		payload, _ = json.Marshal(map[string]any{"users": h.TopicSubscribers(topic)})
	}
	msg, _ := PhoenixEnvelope{Topic: topic, Event: "seats_changed", Payload: payload}.Marshal()
	h.BroadcastToTopic(topic, msg)
}

func broadcastSpectatorsChanged(h *Hub, gameSvc *game.GameService, topic, roomSlug string) {
	var payload []byte
	if gameSvc != nil {
		if gameUI, err := gameSvc.GetGameUI(roomSlug); err == nil && gameUI != nil {
			payload, _ = json.Marshal(map[string]any{"spectators": gameUI.Spectators})
		}
	}
	if payload == nil {
		payload, _ = json.Marshal(map[string]any{"users": h.TopicSubscribers(topic)})
	}
	msg, _ := PhoenixEnvelope{Topic: topic, Event: "spectators_changed", Payload: payload}.Marshal()
	h.BroadcastToTopic(topic, msg)
}

func broadcastGUIUpdate(h *Hub, gameSvc *game.GameService, topic, roomSlug string) {
	if gameSvc == nil {
		return
	}
	gameUI, err := gameSvc.GetGameUI(roomSlug)
	if err != nil || gameUI == nil {
		return
	}
	payload, err := json.Marshal(gameUI)
	if err != nil {
		return
	}
	msg, _ := PhoenixEnvelope{Topic: topic, Event: "gui_update", Payload: payload}.Marshal()
	h.BroadcastToTopic(topic, msg)
}
