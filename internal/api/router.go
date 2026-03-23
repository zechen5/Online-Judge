// Package api defines the external HTTP surface of the service.
// api 包定义服务对外暴露的 HTTP 接口层。
package api

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"online-judge-backend/internal/api/handlers"
	"online-judge-backend/internal/api/middleware"
	"online-judge-backend/internal/repositories"
	"online-judge-backend/internal/services"
)

func NewServer(db *gorm.DB, judgeQueue services.JudgeDispatcher, judgerStatus string, authService *services.AuthService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	problemRepo := repositories.NewProblemRepository(db)
	submissionRepo := repositories.NewSubmissionRepository(db)
	userRepo := repositories.NewUserRepository(db)

	problemService := services.NewProblemService(problemRepo)
	submissionService := services.NewSubmissionService(submissionRepo, problemRepo, userRepo, judgeQueue)
	statusService := services.NewStatusService(problemRepo, submissionRepo, judgeQueue, judgerStatus)

	authHandler := handlers.NewAuthHandler(authService)
	problemHandler := handlers.NewProblemHandler(problemService)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)
	adminHandler := handlers.NewAdminHandler(problemService, statusService)

	router := gin.New()
	router.Use(gin.Recovery())

	_, currentFile, _, _ := runtime.Caller(0)
	webRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "frontend")

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerStaticRoutes(router, webRoot)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		v1.GET("/problems", problemHandler.ListProblems)
		v1.GET("/problems/:id", problemHandler.GetProblem)
		v1.GET("/problems/:id/rank", submissionHandler.GetLeaderboard)
		v1.GET("/submissions/:id", submissionHandler.GetSubmission)

		authed := v1.Group("")
		authed.Use(middleware.RequireAuth(authService))
		{
			authed.GET("/auth/me", authHandler.Me)
			authed.POST("/problems", problemHandler.CreateProblem)
			authed.GET("/submissions", submissionHandler.ListSubmissions)
			authed.POST("/submissions", submissionHandler.CreateSubmission)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.RequireAuth(authService), middleware.RequireAdmin())
		{
			admin.GET("/problems", adminHandler.ListProblems)
			admin.GET("/problems/:id", adminHandler.GetProblem)
			admin.POST("/problems", adminHandler.CreateProblem)
			admin.PUT("/problems/:id", adminHandler.UpdateProblem)
			admin.PATCH("/problems/:id/status", adminHandler.UpdateProblemStatus)
			admin.DELETE("/problems/:id", adminHandler.DeleteProblem)
			admin.GET("/status", adminHandler.GetStatus)
		}
	}

	return router
}
