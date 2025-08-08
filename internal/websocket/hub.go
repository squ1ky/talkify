package websocket

import (
	"github.com/squ1ky/talkify/internal/models"
	"github.com/squ1ky/talkify/internal/services"
	"log"
)

// Hub manages all WebSocket connections and message routing
type Hub struct {
	clients        map[int]*Client
	Register       chan *Client
	Unregister     chan *Client
	HandleMessage  chan *MessageRequest
	messageService *services.MessageService
}

// NewHub creates a new Hub instance
func NewHub(messageService *services.MessageService) *Hub {
	return &Hub{
		clients:        make(map[int]*Client),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		HandleMessage:  make(chan *MessageRequest),
		messageService: messageService,
	}
}

// Run starts the Hub and handles all channel operations
// This should be called in a goroutine: go Hub.Run()
func (h *Hub) Run() {
	log.Println("Websocket Hub started")

	for {
		select {
		case client := <-h.Register: // Client connected
			h.clients[client.UserID] = client
			log.Printf("User %d (%s) connected to WebSocket", client.UserID, client.Username)

		case client := <-h.Unregister: // Client disconnected
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
				log.Printf("User %d (%s) disconnected from WebSocket", client.UserID, client.Username)
			}

		case messageReq := <-h.HandleMessage: // Handle incoming message from client
			h.processMessage(messageReq)
		}
	}
}

// processMessage handles message creation and delivery
func (h *Hub) processMessage(req *MessageRequest) {
	createReq := models.MessageCreateRequest{
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	}

	messageResp, err := h.messageService.SendMessage(req.SenderID, createReq)
	if err != nil {
		if senderClient, ok := h.clients[req.SenderID]; ok {
			senderClient.sendError("Failed to Send message: " + err.Error())
		}
		log.Printf("Failed to save message from user %d: %v", req.SenderID, err)
		return
	}

	if receiverClient, isOnline := h.clients[req.ReceiverID]; isOnline {
		receiverClient.SendMessage(messageResp)
	}

	if senderClient, exists := h.clients[req.SenderID]; exists {
		senderClient.SendMessage(messageResp)
	}
}

// BroadcastMessage sends a message to specific user if they're online
// This can be called from outside (e.g., REST API, Kafka consumer)
func (h *Hub) BroadcastMessage(userID int, message *models.MessageResponse) {
	if client, exists := h.clients[userID]; exists {
		client.SendMessage(message)
	}
}

// GetOnlineUsers returns slice of currently connected user IDs
func (h *Hub) GetOnlineUsers() []int {
	userIDs := make([]int, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// IsUserOnline checks if specific user is connected
func (h *Hub) IsUserOnline(userID int) bool {
	_, ok := h.clients[userID]
	return ok
}
