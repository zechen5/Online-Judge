package main

import (
	"log"

	"online-judge-backend/login-flow-subset/internal/api"
	"online-judge-backend/login-flow-subset/internal/config"
	"online-judge-backend/login-flow-subset/internal/repositories"
	"online-judge-backend/login-flow-subset/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := repositories.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	authService := services.NewAuthService(repositories.NewUserRepository(db), cfg.Auth.JWTSecret)
	server := api.NewServer(authService)

	log.Printf("login-flow-subset listening on :%s", cfg.Server.Port)
	if err := server.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
