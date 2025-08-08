package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/squ1ky/talkify/internal/models"
	"log"
	"time"
)

// Client represents a single WebSocket connection from a user
type Client struct {
	Conn     *websocket.Conn
	UserID   int
	Username string
	Send     chan []byte
	Hub      *Hub
}

// IncomingMessage represents message received from client's browser
type IncomingMessage struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	ReceiverID int    `json:"receiver_id"`
}

// OutgoingMessage represents message sent to client's browser
type OutgoingMessage struct {
	Type      string                  `json:"type"`
	Message   *models.MessageResponse `json:"message,omitempty"`
	Error     string                  `json:"error,omitempty"`
	Timestamp time.Time               `json:"timestamp,omitempty"`
}

// MessageRequest represents a message that needs to be processed by Hub
type MessageRequest struct {
	SenderID   int
	ReceiverID int
	Content    string
}

// ReadPump reads messages from the WebSocket connection
// Runs in its own goroutine
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket error for user %d: %v", c.UserID, err)
			break
		}

		var incomingMsg IncomingMessage
		if err := json.Unmarshal(messageBytes, &incomingMsg); err != nil {
			log.Printf("Invalid JSON from user %d: %v", c.UserID, err)
			c.sendError("Invalid message format")
			continue
		}

		c.handleIncomingMessage(incomingMsg)
	}
}

// WritePump sends messages to the WebSocket connection
// Runs in its own goroutine
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Hub closed the chan
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message to browser
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error for user %d: %v", c.UserID, err)
				return
			}
		}

	}
}

// handleIncomingMessage handles incoming message from client
func (c *Client) handleIncomingMessage(msg IncomingMessage) {
	switch msg.Type {
	case "message":
		c.Hub.HandleMessage <- &MessageRequest{
			SenderID:   c.UserID,
			ReceiverID: msg.ReceiverID,
			Content:    msg.Content,
		}
	default:
		c.sendError("Unknown message type: " + msg.Type)
	}
}

// sendError sends error message to client
func (c *Client) sendError(errMsg string) {
	outgoingMsg := OutgoingMessage{
		Type:      "error",
		Error:     errMsg,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(outgoingMsg)
	select {
	case c.Send <- data:
		// Message sent successfully
	default:
		// Channel is full, close connection
		close(c.Send)
	}
}

// SendMessage sends message to client
func (c *Client) SendMessage(message *models.MessageResponse) {
	outgoingMsg := OutgoingMessage{
		Type:      "message",
		Message:   message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(outgoingMsg)
	if err != nil {
		log.Printf("Failed to marshal message for user %d: %v", c.UserID, err)
		return
	}

	select {
	case c.Send <- data:
		// Message sent successfully
	default:
		// Channel is full, close connection
		close(c.Send)
	}
}
