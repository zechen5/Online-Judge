// Package handlers adapts HTTP requests and responses to service-layer calls.
// handlers 包负责把 HTTP 请求与响应适配到服务层调用上。
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/services"
)

type SubmissionHandler struct {
	service *services.SubmissionService
}

func NewSubmissionHandler(service *services.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{service: service}
}

func (h *SubmissionHandler) CreateSubmission(c *gin.Context) {
	var req services.CreateSubmissionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user := CurrentUser(c); user != nil {
		req.UserID = user.ID
	}

	submission, err := h.service.Create(req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrJudgeQueueUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, submission)
}

func (h *SubmissionHandler) GetSubmission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	submission, err := h.service.Get(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if h.service.IsNotFound(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, submission)
}

func (h *SubmissionHandler) ListSubmissions(c *gin.Context) {
	user := CurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.service.ListRecent(user.ID, 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SubmissionHandler) GetLeaderboard(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	entries, err := h.service.GetLeaderboard(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problem_id": uint(id),
		"ranks":      entries,
	})
}
