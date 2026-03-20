package ws

import "encoding/json"

type PhoenixEnvelope struct {
	Topic   string          `json:"topic"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Ref     string          `json:"ref,omitempty"`
	JoinRef string          `json:"join_ref,omitempty"`
}

func (e PhoenixEnvelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
