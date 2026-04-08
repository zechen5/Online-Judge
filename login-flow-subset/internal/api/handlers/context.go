package handlers

import (
	"github.com/gin-gonic/gin"

	"online-judge-backend/login-flow-subset/internal/services"
)

const currentUserKey = "current_user"

func SetCurrentUser(c *gin.Context, user *services.AuthUser) {
	c.Set(currentUserKey, user)
}

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
