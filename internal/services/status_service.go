// Package services contains business rules and cross-repository orchestration.
// services 包承载业务规则以及跨仓储的编排逻辑。
package services

import "online-judge-backend/internal/repositories"

type StatusService struct {
	problems     *repositories.ProblemRepository
	submissions  *repositories.SubmissionRepository
	queue        JudgeDispatcher
	judgerStatus string
}

// SystemStatus is the admin-facing summary of backend activity.
// SystemStatus 是面向管理员的后端运行概览。
type SystemStatus struct {
	ProblemCount    int64  `json:"problem_count"`
	SubmissionCount int64  `json:"submission_count"`
	PendingJudges   int64  `json:"pending_judges"`
	JudgerStatus    string `json:"judger_status"`
	ConcurrentJobs  int    `json:"concurrent_jobs"`
	QueueDepth      int    `json:"queue_depth"`
	QueueCapacity   int    `json:"queue_capacity"`
	WorkerCount     int    `json:"worker_count"`
}

// NewStatusService constructs the status aggregation service.
// NewStatusService 构造系统状态聚合服务。
func NewStatusService(problems *repositories.ProblemRepository, submissions *repositories.SubmissionRepository, queue JudgeDispatcher, judgerStatus string) *StatusService {
	return &StatusService{
		problems:     problems,
		submissions:  submissions,
		queue:        queue,
		judgerStatus: judgerStatus,
	}
}

// Get aggregates lightweight counts for the admin monitoring endpoint.
// Get 为管理员监控接口聚合轻量级统计信息。
func (s *StatusService) Get() (SystemStatus, error) {
	problemCount, err := s.problems.Count()
	if err != nil {
		return SystemStatus{}, err
	}
	submissionCount, err := s.submissions.CountAll()
	if err != nil {
		return SystemStatus{}, err
	}
	pendingCount, err := s.submissions.CountPending()
	if err != nil {
		return SystemStatus{}, err
	}

	stats := s.queue.Stats()
	return SystemStatus{
		ProblemCount:    problemCount,
		SubmissionCount: submissionCount,
		PendingJudges:   pendingCount,
		JudgerStatus:    s.judgerStatus,
		ConcurrentJobs:  stats.BusyWorkers,
		QueueDepth:      stats.QueueDepth,
		QueueCapacity:   stats.QueueCapacity,
		WorkerCount:     stats.WorkerCount,
	}, nil
}
