// Package main is the application entrypoint; it wires configuration, storage,
// the asynchronous judger, and the HTTP server together.
// main 包是应用入口；它负责把配置、存储、异步判题链路和 HTTP 服务组装起来。
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
	cfg := config.Load()

	db, err := repositories.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	problemRepo := repositories.NewProblemRepository(db)
	submissionRepo := repositories.NewSubmissionRepository(db)

	evaluator, judgerStatus, err := judger.BuildEvaluator(cfg.Judge)
	if err != nil {
		log.Printf("judge evaluator setup failed, falling back to local executor: %v", err)
		evaluator = judger.NewLocalEvaluator()
		judgerStatus = "async-local-fallback"
	}

	judgeService := services.NewJudgeService(evaluator)
	judgeProcessor := services.NewSubmissionJudgeProcessor(submissionRepo, problemRepo, judgeService)
	judgeQueue := judger.NewAsyncQueue(judgeProcessor, cfg.Judge.WorkerCount, cfg.Judge.QueueCapacity)

	authService := services.NewAuthService(repositories.NewUserRepository(db), cfg.Auth.JWTSecret)
	server := api.NewServer(db, judgeQueue, judgerStatus, authService)

	log.Printf("judge backend mode: %s", judgerStatus)
	if err := server.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
