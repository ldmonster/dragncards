package ws

import (
	"errors"
	"strings"

	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
)

// DispatchChannelMessage routes an incoming envelope to the correct channel handler.
// senderConn is the originating *Conn (nil on offline POST path).
func DispatchChannelMessage(h *Hub, roomSvc *room.RoomService, gameSvc *game.GameService, replaySvc *replay.ReplayService, lfgSvc *lfg.LfgService, clientID string, env PhoenixEnvelope, senderConn *Conn) (*PhoenixEnvelope, error) {
	if env.Topic == "" {
		return nil, errors.New("missing topic")
	}
	if env.Event == "" {
		return nil, errors.New("missing event")
	}

	switch {
	case IsRoomTopic(env.Topic):
		return HandleRoomChannel(h, roomSvc, gameSvc, replaySvc, clientID, env, senderConn)
	case IsChatTopic(env.Topic):
		return HandleChatChannel(h, clientID, env)
	case IsLobbyTopic(env.Topic):
		return HandleLobbyChannel(h, clientID, env)
	case IsLfgTopic(env.Topic):
		return HandleLfgChannel(h, lfgSvc, clientID, env)
	case IsMyTopic(env.Topic):
		return HandleMyTopicChannel(h, clientID, env)
	default:
		return nil, errors.New("unknown channel")
	}
}

func TopicFromString(topic string) string {
	return strings.TrimSpace(topic)
}
