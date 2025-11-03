package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"Go-Lang-project-01/pkg/logger"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	EventUserCreated      EventType = "user.created"
	EventUserUpdated      EventType = "user.updated"
	EventUserDeleted      EventType = "user.deleted"
	EventUserRoleChanged  EventType = "user.role.changed"
	EventProfileUpdated   EventType = "profile.updated"
	EventPasswordChanged  EventType = "password.changed"
	EventSystemAlert      EventType = "system.alert"
	EventHealthStatusChanged EventType = "health.status.changed"
)

// Message represents a WebSocket message
type Message struct {
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID uint
	Role   string
	Hub    *Hub
	Conn   *Conn
	Send   chan Message
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Broadcast messages to all clients
	Broadcast chan Message

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message, 256), // Buffered channel
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logger.Info("WebSocket client connected",
				"client_id", client.ID,
				"user_id", client.UserID,
				"total_clients", len(h.clients),
			)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				logger.Info("WebSocket client disconnected",
					"client_id", client.ID,
					"user_id", client.UserID,
					"total_clients", len(h.clients),
				)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, close and unregister
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.Send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
					logger.Warn("Client send channel full, disconnecting",
						"client_id", client.ID,
					)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(eventType EventType, data map[string]interface{}) {
	message := Message{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}
	
	select {
	case h.Broadcast <- message:
		logger.Debug("Broadcasting message", "type", eventType, "clients", len(h.clients))
	default:
		logger.Warn("Broadcast channel full, message dropped", "type", eventType)
	}
}

// BroadcastToUser sends a message to a specific user (all their connections)
func (h *Hub) BroadcastToUser(userID uint, eventType EventType, data map[string]interface{}) {
	message := Message{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
				count++
			default:
				logger.Warn("Client send channel full", "client_id", client.ID)
			}
		}
	}

	if count > 0 {
		logger.Debug("Message sent to user", "user_id", userID, "connections", count, "type", eventType)
	}
}

// BroadcastToRole sends a message to all users with a specific role
func (h *Hub) BroadcastToRole(role string, eventType EventType, data map[string]interface{}) {
	message := Message{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.Role == role {
			select {
			case client.Send <- message:
				count++
			default:
				logger.Warn("Client send channel full", "client_id", client.ID)
			}
		}
	}

	if count > 0 {
		logger.Debug("Message sent to role", "role", role, "clients", count, "type", eventType)
	}
}

// GetStats returns current hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients": len(h.clients),
		"timestamp":     time.Now(),
	}

	// Count by role
	roleCount := make(map[string]int)
	for client := range h.clients {
		roleCount[client.Role]++
	}
	stats["by_role"] = roleCount

	return stats
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(CloseMessage, []byte{})
				return
			}

			// Send JSON message
			data, err := json.Marshal(message)
			if err != nil {
				logger.Error("Failed to marshal message", "error", err)
				continue
			}

			if err := c.Conn.WriteMessage(TextMessage, data); err != nil {
				logger.Error("Write error", "error", err, "client_id", c.ID)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if IsUnexpectedCloseError(err, CloseGoingAway, CloseAbnormalClosure) {
				logger.Error("WebSocket read error", "error", err, "client_id", c.ID)
			}
			break
		}
		// Currently we don't process incoming messages from clients
		// This is a broadcast-only system
	}
}
