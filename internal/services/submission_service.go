// Package service contains business rules and cross-repository orchestration.
// service 包承载业务规则以及跨仓储的编排逻辑。
package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/repositories"
)

type SubmissionService struct {
	submissions *repositories.SubmissionRepository
	problems    *repositories.ProblemRepository
	users       *repositories.UserRepository
	judge       *JudgeService
}

// CreateSubmissionInput is the payload accepted from the HTTP layer.
// CreateSubmissionInput 是从 HTTP 层进入服务层的提交载荷。
type CreateSubmissionInput struct {
	UserID    uint   `json:"user_id"`
	ProblemID uint   `json:"problem_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

// RankEntry is the API-facing leaderboard row.
// RankEntry 是面向 API 的排行榜条目结构。
type RankEntry struct {
	Rank        int       `json:"rank"`
	Username    string    `json:"username"`
	Runtime     int       `json:"runtime"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// NewSubmissionService builds the submission orchestration service.
// NewSubmissionService 构建提交编排服务。
func NewSubmissionService(submissions *repositories.SubmissionRepository, problems *repositories.ProblemRepository, users *repositories.UserRepository, judge *JudgeService) *SubmissionService {
	return &SubmissionService{
		submissions: submissions,
		problems:    problems,
		users:       users,
		judge:       judge,
	}
}

// Create validates a submission, ensures references exist, persists it, then
// synchronously asks the current evaluator for a judge result.
// Create 会校验提交、确保关联对象存在、先持久化记录，再同步调用当前判题实现获取结果。
func (s *SubmissionService) Create(input CreateSubmissionInput) (*models.Submission, error) {
	if input.UserID == 0 || input.ProblemID == 0 {
		return nil, errors.New("user_id and problem_id are required")
	}
	if input.Language == "" || input.Code == "" {
		return nil, errors.New("language and code are required")
	}

	// Only published problems can receive submissions.
	// 只有已发布题目才允许提交。
	problem, err := s.problems.GetByID(input.ProblemID)
	if err != nil {
		return nil, err
	}
	if problem.Status != models.ProblemStatusPublished {
		return nil, errors.New("problem is not published")
	}

	// Placeholder users keep the flow usable before a real auth system exists.
	// 在真实鉴权系统接入前，占位用户机制保证流程可用。
	if _, err := s.users.EnsureUser(input.UserID); err != nil {
		return nil, err
	}

	// Persist a Pending record first so the submission has an ID even if judging fails.
	// 先持久化 Pending 记录，这样即使判题失败，提交也已有可追踪的 ID。
	submission := &models.Submission{
		UserID:    input.UserID,
		ProblemID: input.ProblemID,
		Language:  input.Language,
		Code:      input.Code,
		Status:    models.SubmissionStatusPending,
	}
	if err := s.submissions.Create(submission); err != nil {
		return nil, err
	}

	// Delegate to the injected judger interface; today it is a stub returning AC.
	// 通过注入的判题接口执行判题；当前实现是固定返回 AC 的 stub。
	result, err := s.judge.Evaluate(problem, submission, problem.TestCases)
	if err != nil {
		// Convert evaluator failures into an RE-like stored result for visibility.
		// 将判题器内部错误转成类似 RE 的持久化结果，便于接口可见。
		submission.Status = models.SubmissionStatusRE
		submission.ErrorMsg = err.Error()
	} else {
		submission.Status = result.Status
		submission.Runtime = result.Runtime
		submission.Memory = result.Memory
		submission.ErrorMsg = result.ErrorMsg
	}

	// Save the final judge outcome back onto the same submission row.
	// 将最终判题结果回写到同一条提交记录。
	if err := s.submissions.Update(submission); err != nil {
		return nil, err
	}
	return s.submissions.GetByID(submission.ID)
}

// Get returns a submission by ID.
// Get 按 ID 返回提交记录。
func (s *SubmissionService) Get(id uint) (*models.Submission, error) {
	return s.submissions.GetByID(id)
}

// GetLeaderboard returns the top three accepted submissions ordered by runtime.
// GetLeaderboard 返回按运行时间排序的前三名通过提交。
func (s *SubmissionService) GetLeaderboard(problemID uint) ([]RankEntry, error) {
	submissions, err := s.submissions.TopRank(problemID, 3)
	if err != nil {
		return nil, err
	}

	// Rank numbers are assigned in memory after the database ordering is applied.
	// 排名数字在数据库排序完成后，于内存中顺序生成。
	entries := make([]RankEntry, 0, len(submissions))
	for i, submission := range submissions {
		entries = append(entries, RankEntry{
			Rank:        i + 1,
			Username:    submission.User.Username,
			Runtime:     submission.Runtime,
			SubmittedAt: submission.CreatedAt,
		})
	}
	return entries, nil
}

// IsNotFound lets handlers map repository misses to HTTP 404.
// IsNotFound 让处理器可以把仓储层未命中映射成 HTTP 404。
func (s *SubmissionService) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
