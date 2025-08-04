package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/talkify/internal/middleware"
	"github.com/squ1ky/talkify/internal/models"
	"github.com/squ1ky/talkify/internal/services"
	"net/http"
)

// SetupRouter initializes gin.Engine with routes and middleware
func SetupRouter(cfgSecret string, userService *services.UserService, messageService *services.MessageService) *gin.Engine {
	r := gin.Default()

	apiV1 := r.Group("/api/v1")

	// Public routes
	apiV1.POST("/auth/register", func(c *gin.Context) {
		var req models.UserCreateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		userResp, err := userService.Register(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, userResp)
	})

	apiV1.POST("/auth/login", func(c *gin.Context) {
		var req models.UserLoginRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		user, err := userService.Login(req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		// TODO: add jwt generation here

		c.JSON(http.StatusOK, gin.H{
			"message": "login success",
			"user_id": user.ID,
		})
	})

	auth := apiV1.Group("/")
	auth.Use(middleware.JWTMiddleware(cfgSecret))

	auth.GET("/users", func(c *gin.Context) {
		users, total, err := userService.List(50, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get users",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"users": users,
		})
	})

	// TODO: add another routes

	return r
}
