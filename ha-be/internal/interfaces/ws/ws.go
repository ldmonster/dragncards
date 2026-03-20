package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WSHandler struct {
	hub       *Hub
	roomSvc   *room.RoomService
	lfgSvc    *lfg.LfgService
	gameSvc   *game.GameService
	replaySvc *replay.ReplayService
}

func NewWSHandler(hub *Hub, roomSvc *room.RoomService, lfgSvc *lfg.LfgService, gameSvc *game.GameService, replaySvc *replay.ReplayService) *WSHandler {
	return &WSHandler{hub: hub, roomSvc: roomSvc, lfgSvc: lfgSvc, gameSvc: gameSvc, replaySvc: replaySvc}
}

func writeError(c *Conn, topic, ref, msg string) {
	if c == nil {
		return
	}
	errEnv := PhoenixEnvelope{Topic: topic, Event: "error", Payload: json.RawMessage(`{"error":"` + msg + `"}`), Ref: ref}
	if b, err := errEnv.Marshal(); err == nil {
		_ = c.Send(b)
	}
}

func (w *WSHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		token := strings.TrimSpace(req.URL.Query().Get("token"))
		if token == "" {
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write([]byte(`{"error":"token required"}`))
			return
		}

		userID, err := auth.ParseToken(token)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write([]byte(`{"error":"invalid token"}`))
			return
		}

		conn, err := websocket.Accept(rw, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}

		clientID := userID
		if clientID == "" {
			clientID = "anon-" + time.Now().Format("20060102150405.000000000")
		}

		if _, exists := w.hub.clients[clientID]; exists {
			clientID = clientID + "-" + time.Now().Format("20060102150405.000000000")
		}

		c := NewConn(conn)
		w.hub.Register(clientID, c)
		defer w.hub.Unregister(clientID)

		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		go c.runWritePump(ctx)

		for {
			var env PhoenixEnvelope
			if err := wsjson.Read(ctx, conn, &env); err != nil {
				return
			}

			if env.Topic == "" || env.Event == "" {
				writeError(c, env.Topic, env.Ref, "topic and event required")
				continue
			}

			resp, err := w.DispatchWithSender(clientID, env, c)
			if err != nil {
				writeError(c, env.Topic, env.Ref, err.Error())
				continue
			}
			if resp != nil {
				if b, err := resp.Marshal(); err == nil {
					_ = c.Send(b)
				}
			}
		}
		return
	}

	if req.Method == http.MethodPost {
		// backward compatibility: REST-like enveloping for non-WebSocket clients
		var env PhoenixEnvelope
		if err := json.NewDecoder(req.Body).Decode(&env); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"error":"invalid envelope"}`))
			return
		}

		if env.Topic == "" || env.Event == "" {
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = rw.Write([]byte(`{"error":"topic and event required"}`))
			return
		}

		id := req.Header.Get("X-Client-ID")
		if id == "" {
			id = "offline-" + time.Now().Format("20060102150405.000000000")
		}

		writeJSON := func(status int, body []byte) {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(status)
			_, _ = rw.Write(body)
		}

		if env.Event == "phx_join" {
			w.hub.Subscribe(env.Topic, id, NewConn(nil))
			writeJSON(http.StatusOK, []byte(`{"status":"joined", "client_id":"`+id+`"}`))
			return
		}

		if env.Event == "phx_leave" {
			w.hub.Unsubscribe(env.Topic, id)
			writeJSON(http.StatusOK, []byte(`{"status":"left", "client_id":"`+id+`"}`))
			return
		}

		if env.Event == "broadcast" {
			w.hub.BroadcastToTopic(env.Topic, env.Payload)
			writeJSON(http.StatusOK, []byte(`{"status":"broadcast", "client_id":"`+id+`"}`))
			return
		}

		if env.Event == "game_action" {
			_, err := w.DispatchChannelMessage(id, env)
			if err != nil {
				writeError(nil, env.Topic, env.Ref, err.Error())
				writeJSON(http.StatusBadRequest, []byte(`{"error":"`+err.Error()+`"}`))
				return
			}
			writeJSON(http.StatusOK, []byte(`{"status":"sent", "client_id":"`+id+`"}`))
			return
		}

		resp, err := w.DispatchChannelMessage(id, env)
		if err != nil {
			writeError(nil, env.Topic, env.Ref, err.Error())
			writeJSON(http.StatusBadRequest, []byte(`{"error":"`+err.Error()+`"}`))
			return
		}

		if resp != nil {
			if b, err := resp.Marshal(); err == nil {
				writeJSON(http.StatusOK, b)
				return
			}
		}

		writeJSON(http.StatusOK, []byte(`{"status":"ok", "client_id":"`+id+`"}`))
		return
	}

	rw.WriteHeader(http.StatusMethodNotAllowed)
}

func (w *WSHandler) DispatchChannelMessage(clientID string, env PhoenixEnvelope) (*PhoenixEnvelope, error) {
	return DispatchChannelMessage(w.hub, w.roomSvc, w.gameSvc, w.replaySvc, w.lfgSvc, clientID, env, nil)
}

// DispatchWithSender is like DispatchChannelMessage but includes the sender *Conn
// so channels can immediately write a reply before broadcasting.
func (w *WSHandler) DispatchWithSender(clientID string, env PhoenixEnvelope, sender *Conn) (*PhoenixEnvelope, error) {
	return DispatchChannelMessage(w.hub, w.roomSvc, w.gameSvc, w.replaySvc, w.lfgSvc, clientID, env, sender)
}

func RegisterRoutes(mux *http.ServeMux, handler *WSHandler) {
	mux.Handle("/be/socket", handler)
}
