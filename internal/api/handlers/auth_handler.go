// Package handler adapts HTTP requests and responses to service-layer calls.
// handler 包负责把 HTTP 请求与响应适配到服务层调用上。
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/services"
)

// AuthHandler exposes login/register/profile endpoints.
// AuthHandler 暴露登录、注册和当前用户资料接口。
type AuthHandler struct {
	service *services.AuthService
}

// NewAuthHandler constructs the auth HTTP adapter.
// NewAuthHandler 构造认证相关的 HTTP 适配器。
func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register creates a new user account.
// Register 创建一个新的用户账号。
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Login authenticates a user and returns a token.
// Login 校验用户身份并返回令牌。
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Me returns the authenticated user's latest stored profile.
// Me 返回当前已认证用户的最新资料。
func (h *AuthHandler) Me(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	current, err := h.service.GetCurrentUser(user.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": current})
}
