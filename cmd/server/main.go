// Package main is the application entrypoint; it wires configuration, storage,
// the judger stub, and the HTTP server together.
// main 包是应用入口；它负责把配置、存储、判题桩实现和 HTTP 服务组装起来。
package main

import (
	"log"

	"online-judge-backend/internal/api"
	"online-judge-backend/internal/config"
	"online-judge-backend/internal/judger"
	"online-judge-backend/internal/repositories"
	"online-judge-backend/internal/services"
)

func main() {
	// Load runtime settings from environment variables.
	// 从环境变量加载运行时配置。
	cfg := config.Load()

	// Initialize the database connection and auto-migrate core tables.
	// 初始化数据库连接并自动迁移核心数据表。
	db, err := repositories.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	// Inject the stub judger so submission flow is complete before the real
	// sandboxed judger exists.
	// 注入判题桩实现，这样在真实沙箱判题器上线前，提交流程也能完整打通。
	judgeService := services.NewJudgeService(judger.NewStub())
	authService := services.NewAuthService(repositories.NewUserRepository(db), cfg.Auth.JWTSecret)
	server := api.NewServer(db, judgeService, authService)

	// Start the HTTP server on the configured port.
	// 在配置指定的端口启动 HTTP 服务。
	if err := server.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
