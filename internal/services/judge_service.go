// Package services contains business rules and cross-repository orchestration.
// services 包承载业务规则以及跨仓储的编排逻辑。
package services

import (
	"context"

	"online-judge-backend/internal/models"
)

// JudgeResult is the normalized output shape returned by any judger backend.
// JudgeResult 是任意判题后端都应返回的统一结果结构。
type JudgeResult struct {
	Status   string
	Runtime  int
	Memory   int
	ErrorMsg string
}

// Evaluator abstracts the concrete judge implementation behind the service layer.
// Evaluator 抽象了具体判题实现，使服务层不依赖底层执行细节。
type Evaluator interface {
	Evaluate(ctx context.Context, problem *models.Problem, submission *models.Submission, testCases []models.TestCase) (JudgeResult, error)
}

// JudgeService is a thin wrapper that keeps the evaluator injectable.
// JudgeService 是一层轻包装，用于保持判题实现可注入。
type JudgeService struct {
	evaluator Evaluator
}

// NewJudgeService binds a concrete evaluator implementation.
// NewJudgeService 绑定一个具体的判题实现。
func NewJudgeService(evaluator Evaluator) *JudgeService {
	return &JudgeService{evaluator: evaluator}
}

// Evaluate delegates the judging work to the configured evaluator.
// Evaluate 把判题工作转发给当前配置的判题实现。
func (s *JudgeService) Evaluate(ctx context.Context, problem *models.Problem, submission *models.Submission, testCases []models.TestCase) (JudgeResult, error) {
	return s.evaluator.Evaluate(ctx, problem, submission, testCases)
}
