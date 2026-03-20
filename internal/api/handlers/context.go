// Package handler adapts HTTP requests and responses to service-layer calls.
// handler 包负责把 HTTP 请求与响应适配到服务层调用上。
package handlers

import (
	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/services"
)

const currentUserKey = "current_user"

// SetCurrentUser stores the authenticated user on the request context.
// SetCurrentUser 将已认证用户写入请求上下文。
func SetCurrentUser(c *gin.Context, user *services.AuthUser) {
	c.Set(currentUserKey, user)
}

// CurrentUser returns the authenticated user from the request context.
// CurrentUser 从请求上下文中读取已认证用户。
func CurrentUser(c *gin.Context) *services.AuthUser {
	value, ok := c.Get(currentUserKey)
	if !ok {
		return nil
	}
	user, ok := value.(*services.AuthUser)
	if !ok {
		return nil
	}
	return user
}
