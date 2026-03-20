package ws

import "sync"

type Hub struct {
	mu           sync.RWMutex
	clients      map[string]*Conn
	topicClients map[string]map[string]*Conn
}

func NewHub() *Hub {
	return &Hub{clients: map[string]*Conn{}, topicClients: map[string]map[string]*Conn{}}
}

func (h *Hub) Register(id string, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[id] = c
}

func (h *Hub) Unregister(id string) {
	h.mu.Lock()
	c, ok := h.clients[id]
	if ok {
		_ = c.Close()
		delete(h.clients, id)
	}
	for topic := range h.topicClients {
		delete(h.topicClients[topic], id)
	}
	h.mu.Unlock()
}

func (h *Hub) Subscribe(topic, id string, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.topicClients[topic]; !ok {
		h.topicClients[topic] = make(map[string]*Conn)
	}
	h.topicClients[topic][id] = c
}

func (h *Hub) Unsubscribe(topic, id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.topicClients[topic]; ok {
		delete(clients, id)
	}
}

func (h *Hub) BroadcastToTopic(topic string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.topicClients[topic]; ok {
		for _, c := range clients {
			if c == nil {
				continue
			}
			_ = c.Send(msg)
		}
	}
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if c == nil {
			continue
		}
		_ = c.Send(msg)
	}
}

// TopicSubscribers returns a snapshot of client IDs currently subscribed to a topic.
func (h *Hub) TopicSubscribers(topic string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.topicClients[topic]
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(clients))
	for id := range clients {
		ids = append(ids, id)
	}
	return ids
}
