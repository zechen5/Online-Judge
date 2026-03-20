// Package middleware contains reusable HTTP middleware for the API server.
// middleware 包存放 API 服务复用的 HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/api/handlers"
	"online-judge-backend/internal/models"
	"online-judge-backend/internal/services"
)

// RequireAuth ensures the request carries a valid bearer token.
// RequireAuth 确保请求携带了合法的 Bearer 令牌。
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

// RequireAdmin ensures the authenticated user has admin privileges.
// RequireAdmin 确保当前已认证用户具备管理员权限。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := handlers.CurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if user.Role != models.UserRoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
