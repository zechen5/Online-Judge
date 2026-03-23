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

type ProblemHandler struct {
	service *services.ProblemService
}

// testCasePayload is the request/response shape used at the HTTP layer.
// testCasePayload 是 HTTP 层使用的测试用例请求/响应结构。
type testCasePayload struct {
	Input    string `json:"input"`
	Output   string `json:"output"`
	IsSample bool   `json:"is_sample"`
}

// createProblemRequest is the public/admin JSON contract for creating problems.
// createProblemRequest 是公开接口和管理接口共用的题目创建 JSON 结构。
type createProblemRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TimeLimit   int               `json:"time_limit"`
	MemoryLimit int               `json:"memory_limit"`
	CreatorID   uint              `json:"creator_id"`
	TestCases   []testCasePayload `json:"test_cases"`
}

// NewProblemHandler constructs the problem HTTP adapter.
// NewProblemHandler 构造题目相关的 HTTP 适配器。
func NewProblemHandler(service *services.ProblemService) *ProblemHandler {
	return &ProblemHandler{service: service}
}

// ListProblems handles paginated published-problem queries.
// ListProblems 处理已发布题目的分页查询请求。
func (h *ProblemHandler) ListProblems(c *gin.Context) {
	// Invalid page values fall back to defaults via zero-value handling.
	// 非法分页参数会通过零值处理回退到默认值。
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	problems, total, err := h.service.List(page, pageSize)
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

// GetProblem returns one problem in the response shape required by the spec.
// GetProblem 按需求文档约定的响应结构返回单个题目。
func (h *ProblemHandler) GetProblem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	problem, err := h.service.GetPublished(uint(id))
	if err != nil {
		status := http.StatusInternalServerError
		if h.service.IsNotFound(err) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// Convert only sample cases into the externally visible sample list.
	// 仅把样例用例转换成对外可见的 samples 列表。
	samples := make([]gin.H, 0)
	for _, tc := range problem.TestCases {
		if tc.IsSample {
			samples = append(samples, gin.H{
				"input":  tc.Input,
				"output": tc.Output,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          problem.ID,
		"title":       problem.Title,
		"description": problem.Description,
		"limits": gin.H{
			"time":   problem.TimeLimit,
			"memory": problem.MemoryLimit,
		},
		"samples":    samples,
		"status":     problem.Status,
		"creator_id": problem.CreatorID,
		"created_at": problem.CreatedAt,
	})
}

// CreateProblem accepts user-submitted problems and leaves them pending review.
// CreateProblem 接收用户提交的题目，并将其置为待审核状态。
func (h *ProblemHandler) CreateProblem(c *gin.Context) {
	var req createProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user := CurrentUser(c); user != nil {
		req.CreatorID = user.ID
	}

	// Public problem creation always starts in Pending status.
	// 公开题目创建接口一律以 Pending 状态落库。
	problem, err := h.service.Create(services.CreateProblemInput{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
		CreatorID:   req.CreatorID,
		Status:      models.ProblemStatusPending,
		TestCases:   toModelTestCases(req.TestCases),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, problem)
}

// toModelTestCases converts HTTP payloads into model-layer entities.
// toModelTestCases 把 HTTP 载荷转换成模型层实体。
func toModelTestCases(items []testCasePayload) []models.TestCase {
	testCases := make([]models.TestCase, 0, len(items))
	for i, item := range items {
		// SortOrder starts at 1 so author intent is preserved and human-readable.
		// SortOrder 从 1 开始，既保留作者顺序，也更符合人工理解。
		testCases = append(testCases, models.TestCase{
			Input:     item.Input,
			Output:    item.Output,
			IsSample:  item.IsSample,
			SortOrder: i + 1,
		})
	}
	return testCases
}
