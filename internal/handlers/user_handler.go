package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/talkify/internal/models"
	"github.com/squ1ky/talkify/internal/services"
	"net/http"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
	jwtService  *services.JWTService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, jwtService *services.JWTService) *UserHandler {
	return &UserHandler{userService: userService, jwtService: jwtService}
}

// RegisterPublicRoutes adds public user routes (no auth required)
func (h *UserHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
}

// RegisterProtectedRoutes adds protected user routes (auth required)
func (h *UserHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.GET("/users", h.GetUsers)
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserCreateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userResp, err := h.userService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "user already exists",
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, userResp)
}

// Login handles user authentication
func (h *UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.userService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// GetUsers handles listing all users
func (h *UserHandler) GetUsers(c *gin.Context) {
	limit, offset := parseLimitOffset(c, 50, 0)

	users, total, err := h.userService.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get users",
		})
		return
	}

	c.JSON(http.StatusOK, models.UserListResponse{
		Users: users,
		Total: total,
	})
}
