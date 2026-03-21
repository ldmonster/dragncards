package ws

import (
	"encoding/json"
	"strings"
)

func IsMyTopic(topic string) bool {
	return strings.HasPrefix(topic, "my_topic:")
}

func HandleMyTopicChannel(h *Hub, clientID string, env PhoenixEnvelope) (*PhoenixEnvelope, error) {
	switch env.Event {
	case "phx_join":
		h.Subscribe(env.Topic, clientID, h.Client(clientID))
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil
	case "new_message":
		broadcast := PhoenixEnvelope{Topic: env.Topic, Event: "new_message", Payload: env.Payload}
		payload, _ := broadcast.Marshal()
		h.BroadcastToTopic(env.Topic, payload)
		return nil, nil
	default:
		return nil, nil
	}
}
