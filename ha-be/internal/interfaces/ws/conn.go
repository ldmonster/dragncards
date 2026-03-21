//nolint:staticcheck
package ws

import (
	"context"
	"errors"
	"sync"

	//nolint:staticcheck
	"nhooyr.io/websocket"
)

type Conn struct {
	ws      *websocket.Conn
	send    chan []byte
	closeMu sync.Mutex
	closed  bool
}

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{
		ws:   ws,
		send: make(chan []byte, 32),
	}
}

func (c *Conn) Send(msg []byte) error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	if c.closed {
		return errors.New("connection closed")
	}
	select {
	case c.send <- msg:
		return nil
	default:
		return errors.New("send queue full")
	}
}

func (c *Conn) Close() error {
	c.closeMu.Lock()
	if c.closed {
		c.closeMu.Unlock()
		return nil
	}
	c.closed = true
	close(c.send)
	c.closeMu.Unlock()
	return c.ws.Close(websocket.StatusNormalClosure, "")
}

func (c *Conn) runWritePump(ctx context.Context) {
	for msg := range c.send {
		if msg == nil {
			continue
		}
		if err := c.ws.Write(ctx, websocket.MessageText, msg); err != nil {
			return
		}
	}
}
