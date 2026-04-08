package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"online-judge-backend/login-flow-subset/internal/api/handlers"
	"online-judge-backend/login-flow-subset/internal/services"
)

func RequireAuth(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		user, err := auth.ParseToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		handlers.SetCurrentUser(c, user)
		c.Next()
	}
}
