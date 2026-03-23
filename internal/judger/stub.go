// Package judger contains pluggable judge implementations.
// judger 包存放可插拔的判题实现。
package judger

import (
	"context"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/services"
)

type Stub struct{}

// NewStub returns the temporary evaluator used before the real judger exists.
// NewStub 返回真实判题器实现前使用的临时评测器。
func NewStub() *Stub {
	return &Stub{}
}

// Evaluate satisfies the judge interface while intentionally short-circuiting execution.
// Evaluate 满足判题接口，但会刻意短路真实执行。
func (s *Stub) Evaluate(_ context.Context, _ *models.Problem, submission *models.Submission, _ []models.TestCase) (services.JudgeResult, error) {
	return services.JudgeResult{
		Status:   models.SubmissionStatusAC,
		Runtime:  1 + len(submission.Code)%100,
		Memory:   1024,
		ErrorMsg: "",
	}, nil
}
