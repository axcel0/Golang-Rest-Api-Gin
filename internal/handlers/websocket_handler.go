package handlers

import (
	"net/http"
	"strconv"

	"Go-Lang-project-01/internal/auth"
	ws "Go-Lang-project-01/internal/websocket"
	"Go-Lang-project-01/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub        *ws.Hub
	jwtManager *auth.JWTManager
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, jwtManager *auth.JWTManager) *WebSocketHandler {
	return &WebSocketHandler{
		hub:        hub,
		jwtManager: jwtManager,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
// @Summary WebSocket connection endpoint
// @Description Establish WebSocket connection for real-time updates (JWT token required via query param)
// @Tags websocket
// @Param token query string true "JWT access token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get JWT token from query parameter (can't use headers in browser WebSocket)
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Validate JWT token
	claims, err := h.jwtManager.ValidateToken(tokenString)
	if err != nil {
		logger.Warn("WebSocket auth failed", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// Create client
	client := &ws.Client{
		ID:     uuid.New().String(),
		UserID: claims.UserID,
		Role:   claims.Role,
		Hub:    h.hub,
		Conn:   &ws.Conn{Conn: conn},
		Send:   make(chan ws.Message, 256),
	}

	// Register client
	h.hub.Register <- client

	// Send welcome message
	client.Send <- ws.Message{
		Type: ws.EventType("connection.established"),
		Data: map[string]interface{}{
			"client_id": client.ID,
			"user_id":   client.UserID,
			"role":      client.Role,
			"message":   "Welcome to WebSocket real-time updates",
		},
	}

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// GetStats returns WebSocket hub statistics
// @Summary Get WebSocket statistics
// @Description Get current WebSocket connection statistics (admin only)
// @Tags websocket
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /ws/stats [get]
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	// Get user from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userRole, exists := c.Get("userRole")
	if !exists || (userRole != "admin" && userRole != "superadmin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	stats := h.hub.GetStats()
	stats["user_id"] = userID

	c.JSON(http.StatusOK, stats)
}

// BroadcastMessage sends a test broadcast message (admin only)
// @Summary Broadcast test message
// @Description Send a test message to all connected WebSocket clients (admin only)
// @Tags websocket
// @Security Bearer
// @Accept json
// @Produce json
// @Param message body map[string]interface{} true "Message to broadcast"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Router /ws/broadcast [post]
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	// Check admin access
	userRole, exists := c.Get("userRole")
	if !exists || (userRole != "admin" && userRole != "superadmin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Broadcast message
	h.hub.BroadcastToAll(ws.EventSystemAlert, payload)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "broadcast sent",
		"stats":   h.hub.GetStats(),
	})
}

// NotifyUserUpdate sends user update notification to specific user
func (h *WebSocketHandler) NotifyUserUpdate(userID uint, data map[string]interface{}) {
	h.hub.BroadcastToUser(userID, ws.EventUserUpdated, data)
}

// NotifyUserCreated broadcasts user created event to admins
func (h *WebSocketHandler) NotifyUserCreated(data map[string]interface{}) {
	h.hub.BroadcastToRole("admin", ws.EventUserCreated, data)
	h.hub.BroadcastToRole("superadmin", ws.EventUserCreated, data)
}

// NotifyUserDeleted broadcasts user deleted event to admins
func (h *WebSocketHandler) NotifyUserDeleted(data map[string]interface{}) {
	h.hub.BroadcastToRole("admin", ws.EventUserDeleted, data)
	h.hub.BroadcastToRole("superadmin", ws.EventUserDeleted, data)
}

// NotifyRoleChanged broadcasts role change event
func (h *WebSocketHandler) NotifyRoleChanged(userID uint, data map[string]interface{}) {
	// Notify the user whose role changed
	h.hub.BroadcastToUser(userID, ws.EventUserRoleChanged, data)

	// Notify all admins
	h.hub.BroadcastToRole("admin", ws.EventUserRoleChanged, data)
	h.hub.BroadcastToRole("superadmin", ws.EventUserRoleChanged, data)
}

// NotifyProfileUpdated broadcasts profile update event
func (h *WebSocketHandler) NotifyProfileUpdated(userID uint, data map[string]interface{}) {
	h.hub.BroadcastToUser(userID, ws.EventProfileUpdated, data)
}

// NotifyPasswordChanged broadcasts password change event
func (h *WebSocketHandler) NotifyPasswordChanged(userID uint, data map[string]interface{}) {
	h.hub.BroadcastToUser(userID, ws.EventPasswordChanged, data)
}

// ParseUserID helper to parse user ID from string
func ParseUserID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
