package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/squ1ky/talkify/internal/services"
	ws "github.com/squ1ky/talkify/internal/websocket"
	"log"
	"net/http"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *ws.Hub
	userService *services.UserService
}

// upgrader converts HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, userService *services.UserService) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, userService: userService}
}

// HandleWebSocketConnection upgrades HTTP to WebSocket and manages client
// Route: GET /ws/chat
func (h *WebSocketHandler) HandleWebSocketConnection(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found",
		})
		return
	}

	userID := userIDInterface.(int)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for user %d: %v", userID, err)
		return
	}

	client := &ws.Client{
		Conn:     conn,
		UserID:   userID,
		Username: user.Username,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket connection established for user %d (%s), userID, user.Username")
}

// RegisterRoutes adds WebSocketRoutes to router group
func (h *WebSocketHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ws/chat", h.HandleWebSocketConnection)
}

// GetOnlineUsers returns list of online users for REST API
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	onlineUsers := h.hub.GetOnlineUsers()
	c.JSON(http.StatusOK, gin.H{
		"online_users": onlineUsers,
		"total":        len(onlineUsers),
	})
}
