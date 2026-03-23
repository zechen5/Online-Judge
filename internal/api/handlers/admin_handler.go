// Package handlers adapts HTTP requests and responses to service-layer calls.
// handlers 包负责把 HTTP 请求与响应适配到服务层调用中。
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

// updateProblemRequest is the admin JSON contract for problem review and editing.
// updateProblemRequest 是管理员审核与编辑题目时使用的 JSON 结构。
type updateProblemRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TimeLimit   int               `json:"time_limit"`
	MemoryLimit int               `json:"memory_limit"`
	Status      int               `json:"status"`
	TestCases   []testCasePayload `json:"test_cases"`
}

// updateProblemStatusRequest is the lightweight status-only review payload.
// updateProblemStatusRequest 是仅修改状态时使用的轻量载荷。
type updateProblemStatusRequest struct {
	Status int `json:"status"`
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

// ListProblems returns the full admin review queue, including hidden items.
// ListProblems 返回完整的管理员审核队列，包括隐藏题目。
func (h *AdminHandler) ListProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	problems, total, err := h.problems.ListAll(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     problems,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// GetProblem returns one problem for admin editing, including testcase data.
// GetProblem 返回管理员编辑所需的单个题目详情，包括测试用例。
func (h *AdminHandler) GetProblem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	problem, err := h.problems.Get(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if h.problems.IsNotFound(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, problem)
}

// UpdateProblem rewrites mutable problem fields and testcase data.
// UpdateProblem 会重写题目的可变字段和测试用例数据。
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

// UpdateProblemStatus lets admins approve, hide, or return a problem to pending.
// UpdateProblemStatus 允许管理员发布、隐藏或重新置为待审核。
func (h *AdminHandler) UpdateProblemStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	var req updateProblemStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	problem, err := h.problems.UpdateStatus(uint(id), services.UpdateProblemStatusInput{
		Status: req.Status,
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

// DeleteProblem permanently removes a problem from the review queue.
// DeleteProblem 会把题目从审核队列中永久删除。
func (h *AdminHandler) DeleteProblem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	if err := h.problems.Delete(uint(id)); err != nil {
		status := http.StatusBadRequest
		if h.problems.IsNotFound(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
