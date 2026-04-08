package api

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"

	"online-judge-backend/login-flow-subset/internal/api/handlers"
	"online-judge-backend/login-flow-subset/internal/api/middleware"
	"online-judge-backend/login-flow-subset/internal/services"
)

func NewServer(authService *services.AuthService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	authHandler := handlers.NewAuthHandler(authService)

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

		authed := v1.Group("")
		authed.Use(middleware.RequireAuth(authService))
		{
			authed.GET("/auth/me", authHandler.Me)
		}
	}

	return router
}
