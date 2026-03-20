// Package handler adapts HTTP requests and responses to service-layer calls.
// handler 包负责把 HTTP 请求与响应适配到服务层调用上。
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/services"
)

type AdminHandler struct {
	problems *services.ProblemService
	status   *services.StatusService
}

// updateProblemRequest is the admin JSON contract for problem review/update.
// updateProblemRequest 是管理员审核或更新题目时使用的 JSON 结构。
type updateProblemRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TimeLimit   int               `json:"time_limit"`
	MemoryLimit int               `json:"memory_limit"`
	Status      int               `json:"status"`
	TestCases   []testCasePayload `json:"test_cases"`
}

// NewAdminHandler constructs the admin HTTP adapter.
// NewAdminHandler 构造管理员接口的 HTTP 适配器。
func NewAdminHandler(problems *services.ProblemService, status *services.StatusService) *AdminHandler {
	return &AdminHandler{problems: problems, status: status}
}

// CreateProblem lets admins create already-published problems directly.
// CreateProblem 允许管理员直接创建并发布题目。
func (h *AdminHandler) CreateProblem(c *gin.Context) {
	var req createProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user := CurrentUser(c); user != nil {
		req.CreatorID = user.ID
	}

	// Admin-created problems skip the pending state and become public immediately.
	// 管理员创建的题目跳过待审核阶段，直接变为公开状态。
	problem, err := h.problems.Create(services.CreateProblemInput{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
		CreatorID:   req.CreatorID,
		Status:      models.ProblemStatusPublished,
		TestCases:   toModelTestCases(req.TestCases),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, problem)
}

// UpdateProblem lets admins approve or rewrite a problem and its testcases.
// UpdateProblem 允许管理员审核通过或重写题目及其测试用例。
func (h *AdminHandler) UpdateProblem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	var req updateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	problem, err := h.problems.Update(uint(id), services.UpdateProblemInput{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
		Status:      req.Status,
		TestCases:   toModelTestCases(req.TestCases),
	})
	if err != nil {
		status := http.StatusBadRequest
		if h.problems.IsNotFound(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, problem)
}

// GetStatus exposes a lightweight system summary for admins.
// GetStatus 向管理员暴露轻量级系统状态概览。
func (h *AdminHandler) GetStatus(c *gin.Context) {
	status, err := h.status.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
