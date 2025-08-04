package services

import (
	"errors"
	"github.com/squ1ky/talkify/internal/database"
	"github.com/squ1ky/talkify/internal/models"
	"time"
)

var (
	ErrInvalidContent = errors.New("invalid message content")
	ErrMsgNotFound    = errors.New("message not found")
)

// MessageService manages message-related business logic
type MessageService struct {
	messages *database.MessageRepository
	users    *database.UserRepository
}

// NewMessageService creates new message service
func NewMessageService(messages *database.MessageRepository, users *database.UserRepository) *MessageService {
	return &MessageService{
		messages: messages,
		users:    users,
	}
}

// SendMessage creates new message between two users
func (s *MessageService) SendMessage(senderID int, req models.MessageCreateRequest) (*models.MessageResponse, error) {
	if !models.IsValidMessageContent(req.Content) {
		return nil, ErrInvalidContent
	}
	if _, err := s.users.GetByID(req.ReceiverID); err != nil {
		return nil, errors.New("receiver not found")
	}

	message := models.CreateMessageFromRequest(req, senderID)
	if err := s.messages.Create(message); err != nil {
		return nil, err
	}

	resp := message.ToResponse()
	return &resp, nil
}

// GetConversationHistory returns list of messages between two users with pagination
func (s *MessageService) GetConversationHistory(userID1, userID2, limit, offset int) ([]models.MessageWithUserResponse, int, error) {
	messages, err := s.messages.GetConversationHistory(userID1, userID2, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.messages.CountConversationMessages(userID1, userID2)
	if err != nil {
		return nil, 0, err
	}

	return messages, count, nil
}

// GetRecentConversations returns user's list of recent interlocutors
func (s *MessageService) GetRecentConversations(userID, limit int) ([]models.UserResponse, error) {
	return s.messages.GetRecentConversations(userID, limit)
}

// GetMessagesSince allows to get new messages after some time
func (s *MessageService) GetMessagesSince(userID int, since time.Time) ([]models.MessageWithUserResponse, error) {
	return s.messages.GetMessagesSince(userID, since)
}
