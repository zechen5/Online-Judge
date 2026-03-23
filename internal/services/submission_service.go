// Package services contains business rules and cross-repository orchestration.
// services 包承载业务规则以及跨仓储的编排逻辑。
package services

import (
	"context"
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
	queue       JudgeDispatcher
}

type CreateSubmissionInput struct {
	UserID    uint   `json:"user_id"`
	ProblemID uint   `json:"problem_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

type RankEntry struct {
	Rank        int       `json:"rank"`
	Username    string    `json:"username"`
	Runtime     int       `json:"runtime"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type SubmissionListItem struct {
	ID           uint      `json:"id"`
	ProblemID    uint      `json:"problem_id"`
	ProblemTitle string    `json:"problem_title"`
	Language     string    `json:"language"`
	Status       string    `json:"status"`
	Runtime      int       `json:"runtime"`
	Memory       int       `json:"memory"`
	ErrorMsg     string    `json:"error_msg"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewSubmissionService constructs the submission orchestration service.
// NewSubmissionService 构造提交相关的编排服务。
func NewSubmissionService(submissions *repositories.SubmissionRepository, problems *repositories.ProblemRepository, users *repositories.UserRepository, queue JudgeDispatcher) *SubmissionService {
	return &SubmissionService{
		submissions: submissions,
		problems:    problems,
		users:       users,
		queue:       queue,
	}
}

// Create stores a submission record and schedules async judging.
// Create 会先落库提交记录，再安排异步判题。
func (s *SubmissionService) Create(input CreateSubmissionInput) (*models.Submission, error) {
	if input.UserID == 0 || input.ProblemID == 0 {
		return nil, errors.New("user_id and problem_id are required")
	}
	if input.Language == "" || input.Code == "" {
		return nil, errors.New("language and code are required")
	}

	problem, err := s.problems.GetByID(input.ProblemID)
	if err != nil {
		return nil, err
	}
	if problem.Status != models.ProblemStatusPublished {
		return nil, errors.New("problem is not published")
	}

	if _, err := s.users.EnsureUser(input.UserID); err != nil {
		return nil, err
	}

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

	enqueueCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.queue.Enqueue(enqueueCtx, SubmissionTask{SubmissionID: submission.ID}); err != nil {
		submission.Status = models.SubmissionStatusRE
		submission.ErrorMsg = ErrJudgeQueueUnavailable.Error()
		_ = s.submissions.Update(submission)
		return nil, err
	}

	return s.submissions.GetByID(submission.ID)
}

// Get returns one submission by primary key.
// Get 按主键返回单个提交。
func (s *SubmissionService) Get(id uint) (*models.Submission, error) {
	return s.submissions.GetByID(id)
}

// ListRecent returns recent submissions for one user.
// ListRecent 返回单个用户最近的提交记录。
func (s *SubmissionService) ListRecent(userID uint, limit int) ([]SubmissionListItem, error) {
	if userID == 0 {
		return nil, errors.New("user_id is required")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	submissions, err := s.submissions.ListByUser(userID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]SubmissionListItem, 0, len(submissions))
	for _, submission := range submissions {
		items = append(items, SubmissionListItem{
			ID:           submission.ID,
			ProblemID:    submission.ProblemID,
			ProblemTitle: submission.Problem.Title,
			Language:     submission.Language,
			Status:       submission.Status,
			Runtime:      submission.Runtime,
			Memory:       submission.Memory,
			ErrorMsg:     submission.ErrorMsg,
			CreatedAt:    submission.CreatedAt,
		})
	}
	return items, nil
}

// GetLeaderboard returns the best AC submissions for a problem.
// GetLeaderboard 返回题目的最佳 AC 提交排行。
func (s *SubmissionService) GetLeaderboard(problemID uint) ([]RankEntry, error) {
	submissions, err := s.submissions.TopRank(problemID, 3)
	if err != nil {
		return nil, err
	}

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

// IsNotFound tells handlers whether storage reported a missing row.
// IsNotFound 告诉处理器底层是否返回了记录不存在。
func (s *SubmissionService) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
