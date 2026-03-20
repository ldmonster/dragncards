package ws

import (
	"encoding/json"
)

func IsLobbyTopic(topic string) bool {
	return topic == "lobby:lobby"
}

func HandleLobbyChannel(h *Hub, clientID string, env PhoenixEnvelope) (*PhoenixEnvelope, error) {
	switch env.Event {
	case "phx_join":
		h.Subscribe(env.Topic, clientID, h.clients[clientID])
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil
	case "leave_room":
		h.Unsubscribe(env.Topic, clientID)
		return nil, nil
	case "lobby_update":
		broadcast := PhoenixEnvelope{Topic: env.Topic, Event: "lobby_update", Payload: env.Payload}
		payload, _ := broadcast.Marshal()
		h.BroadcastToTopic(env.Topic, payload)
		return nil, nil
	default:
		return nil, nil
	}
}
