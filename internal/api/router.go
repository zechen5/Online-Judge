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

func NewServer(db *gorm.DB, judgeService *services.JudgeService, authService *services.AuthService) *gin.Engine {
	// Force a quiet production-style Gin mode for this scaffold.
	// 对这个后端骨架强制使用较安静的生产风格 Gin 模式。
	gin.SetMode(gin.ReleaseMode)

	// Build repositories first; services and handlers depend on them.
	// 先构建仓储层；服务层和处理器层都依赖它们。
	problemRepo := repositories.NewProblemRepository(db)
	submissionRepo := repositories.NewSubmissionRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Services hold business rules and orchestration logic.
	// 服务层承载业务规则和编排逻辑。
	problemService := services.NewProblemService(problemRepo)
	submissionService := services.NewSubmissionService(submissionRepo, problemRepo, userRepo, judgeService)
	statusService := services.NewStatusService(problemRepo, submissionRepo)

	// Handlers translate HTTP requests into service calls.
	// 处理器负责把 HTTP 请求转换成服务调用。
	authHandler := handlers.NewAuthHandler(authService)
	problemHandler := handlers.NewProblemHandler(problemService)
	submissionHandler := handlers.NewSubmissionHandler(submissionService)
	adminHandler := handlers.NewAdminHandler(problemService, statusService)

	// Use a minimal middleware set: recovery only.
	// 仅启用最小中间件集合：这里只保留 recovery。
	router := gin.New()
	router.Use(gin.Recovery())

	_, currentFile, _, _ := runtime.Caller(0)
	webRoot := filepath.Join(filepath.Dir(currentFile), "..", "..", "frontend")

	// Health endpoint for smoke checks and container probes.
	// 健康检查接口，便于冒烟测试和容器探针使用。
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerStaticRoutes(router, webRoot)

	// Group all versioned API endpoints under /api/v1.
	// 将所有版本化接口统一挂到 /api/v1 下。
	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		// Problem browsing and user-side problem creation.
		// 题目浏览接口，以及用户侧提交题目接口。
		v1.GET("/problems", problemHandler.ListProblems)
		v1.GET("/problems/:id", problemHandler.GetProblem)
		v1.GET("/problems/:id/rank", submissionHandler.GetLeaderboard)

		// Submission creation and result polling.
		// 提交代码与轮询查看结果接口。
		v1.GET("/submissions/:id", submissionHandler.GetSubmission)

		authed := v1.Group("")
		authed.Use(middleware.RequireAuth(authService))
		{
			authed.GET("/auth/me", authHandler.Me)
			authed.POST("/problems", problemHandler.CreateProblem)
			authed.POST("/submissions", submissionHandler.CreateSubmission)
		}

		// Admin-only routes are grouped for future auth middleware.
		// 管理员路由单独分组，便于后续接入鉴权中间件。
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireAuth(authService), middleware.RequireAdmin())
		{
			admin.POST("/problems", adminHandler.CreateProblem)
			admin.PUT("/problems/:id", adminHandler.UpdateProblem)
			admin.GET("/status", adminHandler.GetStatus)
		}
	}

	return router
}
