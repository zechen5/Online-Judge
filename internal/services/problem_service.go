// Package service contains business rules and cross-repository orchestration.
// service 包承载业务规则以及跨仓储的编排逻辑。
package services

import (
	"errors"

	"gorm.io/gorm"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/repositories"
)

type ProblemService struct {
	repo *repositories.ProblemRepository
}

// CreateProblemInput is the validated service-level payload for problem creation.
// CreateProblemInput 是题目创建流程在服务层使用的输入载荷。
type CreateProblemInput struct {
	Title       string
	Description string
	TimeLimit   int
	MemoryLimit int
	CreatorID   uint
	Status      int
	TestCases   []models.TestCase
}

// UpdateProblemInput is the validated service-level payload for problem updates.
// UpdateProblemInput 是题目更新流程在服务层使用的输入载荷。
type UpdateProblemInput struct {
	Title       string
	Description string
	TimeLimit   int
	MemoryLimit int
	Status      int
	TestCases   []models.TestCase
}

// NewProblemService constructs the problem business service.
// NewProblemService 构造题目业务服务。
func NewProblemService(repo *repositories.ProblemRepository) *ProblemService {
	return &ProblemService{repo: repo}
}

// Create validates business invariants before persisting a problem.
// Create 会先校验业务约束，再持久化题目。
func (s *ProblemService) Create(input CreateProblemInput) (*models.Problem, error) {
	if input.Title == "" {
		return nil, errors.New("title is required")
	}
	if input.TimeLimit <= 0 || input.MemoryLimit <= 0 {
		return nil, errors.New("time_limit and memory_limit must be positive")
	}

	// New testcases are attached directly during initial creation.
	// 新建题目时，测试用例直接作为关联一起写入。
	problem := &models.Problem{
		Title:       input.Title,
		Description: input.Description,
		TimeLimit:   input.TimeLimit,
		MemoryLimit: input.MemoryLimit,
		CreatorID:   input.CreatorID,
		Status:      input.Status,
		TestCases:   input.TestCases,
	}
	if err := s.repo.Create(problem); err != nil {
		return nil, err
	}
	return s.repo.GetByID(problem.ID)
}

// Update rewrites mutable problem fields and replaces testcase data.
// Update 会重写可变题目字段，并替换测试用例数据。
func (s *ProblemService) Update(id uint, input UpdateProblemInput) (*models.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	problem.Title = input.Title
	problem.Description = input.Description
	problem.TimeLimit = input.TimeLimit
	problem.MemoryLimit = input.MemoryLimit
	problem.Status = input.Status

	// Rebind ProblemID for every testcase to keep associations consistent.
	// 为每条测试用例重新绑定 ProblemID，保证关联一致。
	testCases := make([]models.TestCase, 0, len(input.TestCases))
	for _, tc := range input.TestCases {
		tc.ProblemID = problem.ID
		testCases = append(testCases, tc)
	}
	problem.TestCases = testCases

	if err := s.repo.Update(problem); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

// Get returns one problem by primary key.
// Get 按主键返回单个题目。
func (s *ProblemService) Get(id uint) (*models.Problem, error) {
	return s.repo.GetByID(id)
}

// List exposes paginated published problems for public APIs.
// List 为公开接口提供已发布题目的分页结果。
func (s *ProblemService) List(page, pageSize int) ([]models.Problem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.repo.ListPublished((page-1)*pageSize, pageSize)
}

// IsNotFound lets handlers convert storage misses into 404 responses.
// IsNotFound 让处理器可以把存储层未命中转换成 404 响应。
func (s *ProblemService) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
