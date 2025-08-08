package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/talkify/internal/handlers"
	"github.com/squ1ky/talkify/internal/middleware"
	"github.com/squ1ky/talkify/internal/services"
	"github.com/squ1ky/talkify/internal/websocket"
)

// SetupRouter initializes gin.Engine with routes and middleware
func SetupRouter(cfgSecret string, userService *services.UserService, messageService *services.MessageService) *gin.Engine {
	r := gin.Default()

	jwtService := services.NewJWTService(cfgSecret)

	hub := websocket.NewHub(messageService)
	go hub.Run()

	userHandler := handlers.NewUserHandler(userService, jwtService)
	messageHandler := handlers.NewMessageHandler(messageService, userService)
	wsHandler := handlers.NewWebSocketHandler(hub, userService)

	apiV1 := r.Group("/api/v1")

	userHandler.RegisterPublicRoutes(apiV1)

	auth := apiV1.Group("/")
	auth.Use(middleware.JWTMiddleware(cfgSecret))

	userHandler.RegisterProtectedRoutes(auth)
	messageHandler.RegisterProtectedRoutes(auth)
	wsHandler.RegisterRoutes(auth)

	auth.GET("/online-users", wsHandler.GetOnlineUsers)

	return r
}
