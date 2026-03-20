// Package handler adapts HTTP requests and responses to service-layer calls.
// handler 包负责把 HTTP 请求与响应适配到服务层调用上。
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/services"
)

type SubmissionHandler struct {
	service *services.SubmissionService
}

// NewSubmissionHandler constructs the submission HTTP adapter.
// NewSubmissionHandler 构造提交相关的 HTTP 适配器。
func NewSubmissionHandler(service *services.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{service: service}
}

// CreateSubmission receives source code, triggers judging, and returns the row.
// CreateSubmission 接收源码、触发判题，并返回提交记录。
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, submission)
}

// GetSubmission returns the latest stored view of a submission.
// GetSubmission 返回当前存储中的提交结果视图。
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

// GetLeaderboard returns the top-three ranking for a problem.
// GetLeaderboard 返回指定题目的前三排行榜。
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
