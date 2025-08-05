package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/talkify/internal/models"
	"github.com/squ1ky/talkify/internal/services"
	"net/http"
	"strconv"
)

// MessageHandler handles message-related API requests
type MessageHandler struct {
	messages *services.MessageService
	users    *services.UserService
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messages *services.MessageService, users *services.UserService) *MessageHandler {
	return &MessageHandler{messages: messages, users: users}
}

// RegisterProtectedRoutes applies routes on group (/api/v1, secured by JWT-middleware)
func (h *MessageHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.POST("/messages", h.SendMessage)
	rg.GET("/messages/:userID", h.GetConversation)
	rg.GET("/conversations", h.GetConversations)
}

// SendMessage POST /messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user_id not found in context",
		})
		return
	}

	senderID := uid.(int)

	var req models.MessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.messages.SendMessage(senderID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidContent):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to send message",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetConversation GET /messages/:userID
func (h *MessageHandler) GetConversation(c *gin.Context) {
	uid, _ := c.Get("user_id")
	currentID := uid.(int)

	otherStr := c.Param("userID")
	otherID, err := strconv.Atoi(otherStr)
	if err != nil || otherID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid userID",
		})
		return
	}

	limit, offset := parseLimitOffset(c, 50, 0)

	messages, total, err := h.messages.GetConversationHistory(currentID, otherID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get conversation history",
		})
		return
	}

	participant, err := h.users.GetByID(otherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "participant not found",
		})
		return
	}

	resp := models.MessageHistoryResponse{
		Messages:    messages,
		Total:       total,
		Participant: *participant,
	}
	c.JSON(http.StatusOK, resp)
}

// GetConversations GET /conversations
func (h *MessageHandler) GetConversations(c *gin.Context) {
	uid, _ := c.Get("user_id")
	currentID := uid.(int)

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	users, err := h.messages.GetRecentConversations(currentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get conversations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(users),
		"users": users,
	})
}

// parseLimitOffset parses ?limit=&offset=
func parseLimitOffset(c *gin.Context, defLimit, defOffset int) (int, int) {
	limitStr := c.DefaultQuery("limit", strconv.Itoa(defLimit))
	offsetStr := c.DefaultQuery("offset", strconv.Itoa(defOffset))

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = defLimit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset <= 0 {
		offset = defOffset
	}

	return limit, offset
}
