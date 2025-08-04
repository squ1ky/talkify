package server

import (
	"github.com/joho/godotenv"
	"github.com/squ1ky/talkify/internal/config"
	"github.com/squ1ky/talkify/internal/database"
	"github.com/squ1ky/talkify/internal/routers"
	"github.com/squ1ky/talkify/internal/services"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: unable to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	migrationManager, err := database.NewMigrationManager(db)
	if err := migrationManager.Up(); err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	log.Println("Database migrations applied successfully")

	userRepo := database.NewUserRepository(db)
	messageRepo := database.NewMessageRepository(db)
	userService := services.NewUserService(userRepo)
	messageService := services.NewMessageService(messageRepo, userRepo)

	r := routers.SetupRouter(cfg.JWT.Secret, userService, messageService)

	r.Run(cfg.Server.GetServerAddress())
}
