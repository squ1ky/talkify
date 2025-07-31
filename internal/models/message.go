package models

import "time"

// Message represents a message in the system
type Message struct {
	ID         int       `json:"id" db:"id"`
	SenderID   int       `json:"sender_id" db:"sender_id"`
	ReceiverID int       `json:"receiver_id" db:"receiver_id"`
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// MessageCreateRequest represents request for sending a message
type MessageCreateRequest struct {
	ReceiverID int    `json:"receiver_id" binding:"required,min=1"`
	Content    string `json:"content" binding:"required,min=1,max=1000"`
}

// MessageResponse represents message data in API responses
type MessageResponse struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// MessageWithUserResponse represents message with sender/receiver info
type MessageWithUserResponse struct {
	ID        int          `json:"id"`
	Content   string       `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	Sender    UserResponse `json:"sender"`
	Receiver  UserResponse `json:"receiver"`
}

// MessageHistoryResponse represents chat history between two users
type MessageHistoryResponse struct {
	Messages    []MessageWithUserResponse `json:"messages"`
	Total       int                       `json:"total"`
	Participant UserResponse              `json:"participant"`
}

// KafkaMessageEvent represents event published to Kafka
type KafkaMessageEvent struct {
	EventType  string    `json:"event_type"`
	MessageID  int       `json:"message_id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

// ToResponse converts Message to MessageResponse
func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:         m.ID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt,
	}
}

// ToKafkaEvent converts Message to KafkaMessageEvent
func (m *Message) ToKafkaEvent() KafkaMessageEvent {
	return KafkaMessageEvent{
		EventType:  "message.sent",
		MessageID:  m.ID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Content:    m.Content,
		Timestamp:  m.CreatedAt,
	}
}

// CreateMessageFromRequest creates Message from MessageCreateRequest
func CreateMessageFromRequest(req MessageCreateRequest, senderId int) *Message {
	return &Message{
		SenderID:   senderId,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		CreatedAt:  time.Now(),
	}
}

// IsValidMessageContent checks if message content meets requirements
func IsValidMessageContent(content string) bool {
	if len(content) == 0 || len(content) > 1000 {
		return false
	}

	// Check if content is not just whitespace
	trimmed := ""
	for _, char := range content {
		if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
			trimmed += string(char)
		}
	}

	return len(trimmed) > 0
}

// GetChatParticipants returns IDs of chat participants
func (m *Message) GetChatParticipants() []int {
	return []int{m.SenderID, m.ReceiverID}
}
