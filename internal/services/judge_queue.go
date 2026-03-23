// Package services contains business rules and cross-repository orchestration.
// services 包承载业务规则以及跨仓储的编排逻辑。
package services

import (
	"context"
	"errors"
	"fmt"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/repositories"
)

// ErrJudgeQueueUnavailable is returned when the async queue cannot accept work.
// ErrJudgeQueueUnavailable 表示异步判题队列当前无法接收新任务。
var ErrJudgeQueueUnavailable = errors.New("judge queue unavailable")

// SubmissionTask is the queue payload passed from HTTP-facing services to workers.
// SubmissionTask 是从 HTTP 服务层传给后台工作线程的队列载荷。
type SubmissionTask struct {
	SubmissionID uint
}

// JudgeQueueStats captures lightweight runtime metrics for monitoring.
// JudgeQueueStats 描述用于监控的轻量运行时指标。
type JudgeQueueStats struct {
	WorkerCount   int `json:"worker_count"`
	BusyWorkers   int `json:"busy_workers"`
	QueueDepth    int `json:"queue_depth"`
	QueueCapacity int `json:"queue_capacity"`
}

// JudgeDispatcher is the boundary between HTTP submission intake and worker execution.
// JudgeDispatcher 是 HTTP 提交入口与后台工作线程执行之间的边界。
type JudgeDispatcher interface {
	Enqueue(ctx context.Context, task SubmissionTask) error
	Stats() JudgeQueueStats
}

// SubmissionProcessor is the worker-side contract used by async queues.
// SubmissionProcessor 是异步队列在工作线程侧使用的处理契约。
type SubmissionProcessor interface {
	Process(ctx context.Context, task SubmissionTask) error
}

// SubmissionJudgeProcessor resolves queued submissions and writes back verdicts.
// SubmissionJudgeProcessor 负责解析排队中的提交并把评测结果回写数据库。
type SubmissionJudgeProcessor struct {
	submissions *repositories.SubmissionRepository
	problems    *repositories.ProblemRepository
	judge       *JudgeService
}

// NewSubmissionJudgeProcessor constructs the worker-side submission processor.
// NewSubmissionJudgeProcessor 构造工作线程使用的提交处理器。
func NewSubmissionJudgeProcessor(submissions *repositories.SubmissionRepository, problems *repositories.ProblemRepository, judge *JudgeService) *SubmissionJudgeProcessor {
	return &SubmissionJudgeProcessor{
		submissions: submissions,
		problems:    problems,
		judge:       judge,
	}
}

// Process loads the queued submission, evaluates it, and persists the final verdict.
// Process 会读取排队的提交，执行评测，并持久化最终判定结果。
func (p *SubmissionJudgeProcessor) Process(ctx context.Context, task SubmissionTask) error {
	submission, err := p.submissions.GetByID(task.SubmissionID)
	if err != nil {
		return err
	}
	if isTerminalSubmissionStatus(submission.Status) {
		return nil
	}

	submission.Status = models.SubmissionStatusJudging
	submission.ErrorMsg = ""
	submission.Runtime = 0
	submission.Memory = 0
	if err := p.submissions.Update(submission); err != nil {
		return err
	}

	problem, err := p.problems.GetByID(submission.ProblemID)
	if err != nil {
		submission.Status = models.SubmissionStatusRE
		submission.ErrorMsg = fmt.Sprintf("load problem: %v", err)
		_ = p.submissions.Update(submission)
		return err
	}

	result, err := p.judge.Evaluate(ctx, problem, submission, problem.TestCases)
	if err != nil {
		submission.Status = models.SubmissionStatusRE
		submission.ErrorMsg = err.Error()
	} else {
		submission.Status = result.Status
		submission.Runtime = result.Runtime
		submission.Memory = result.Memory
		submission.ErrorMsg = result.ErrorMsg
	}

	return p.submissions.Update(submission)
}

func isTerminalSubmissionStatus(status string) bool {
	switch status {
	case models.SubmissionStatusAC, models.SubmissionStatusWA, models.SubmissionStatusTLE, models.SubmissionStatusRE, models.SubmissionStatusCE:
		return true
	default:
		return false
	}
}
