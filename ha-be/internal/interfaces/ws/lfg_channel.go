package ws

import (
	"encoding/json"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
)

func IsLfgTopic(topic string) bool {
	return strings.HasPrefix(topic, "lfg:")
}

func HandleLfgChannel(h *Hub, svc *lfg.LfgService, clientID string, env PhoenixEnvelope) (*PhoenixEnvelope, error) {
	switch env.Event {
	case "phx_join":
		h.Subscribe(env.Topic, clientID, h.Client(clientID))
		return &PhoenixEnvelope{Topic: env.Topic, Event: "phx_reply", Payload: json.RawMessage(`{"status":"ok"}`), Ref: env.Ref}, nil
	case "new_post":
		if svc != nil {
			var req struct {
				PluginID string `json:"plugin_id"`
				UserID   string `json:"user_id"`
				Text     string `json:"text"`
			}
			_ = json.Unmarshal(env.Payload, &req)
			if req.PluginID != "" && req.UserID != "" && req.Text != "" {
				_, _ = svc.Create(req.PluginID, req.UserID, req.Text)
			}
		}
		broadcast := PhoenixEnvelope{Topic: env.Topic, Event: "new_post", Payload: env.Payload}
		payload, _ := broadcast.Marshal()
		h.BroadcastToTopic(env.Topic, payload)
		return nil, nil
	case "delete_post":
		if svc != nil {
			var req struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(env.Payload, &req)
			if req.ID != "" {
				_ = svc.Delete(req.ID)
			}
		}
		broadcast := PhoenixEnvelope{Topic: env.Topic, Event: "delete_post", Payload: env.Payload}
		payload, _ := broadcast.Marshal()
		h.BroadcastToTopic(env.Topic, payload)
		return nil, nil
	default:
		return nil, nil
	}
}
