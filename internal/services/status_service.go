// Package service contains business rules and cross-repository orchestration.
// service 包承载业务规则以及跨仓储的编排逻辑。
package services

import "online-judge-backend/internal/repositories"

type StatusService struct {
	problems    *repositories.ProblemRepository
	submissions *repositories.SubmissionRepository
}

// SystemStatus is the admin-facing summary of backend activity.
// SystemStatus 是面向管理员的后端运行概览。
type SystemStatus struct {
	ProblemCount    int64  `json:"problem_count"`
	SubmissionCount int64  `json:"submission_count"`
	PendingJudges   int64  `json:"pending_judges"`
	JudgerStatus    string `json:"judger_status"`
	ConcurrentJobs  int    `json:"concurrent_jobs"`
}

// NewStatusService constructs the status aggregation service.
// NewStatusService 构造系统状态聚合服务。
func NewStatusService(problems *repositories.ProblemRepository, submissions *repositories.SubmissionRepository) *StatusService {
	return &StatusService{problems: problems, submissions: submissions}
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

	// The real judger/concurrency accounting does not exist yet, so the values
	// are explicit placeholders instead of pretending to be measured data.
	// 真实判题器和并发统计尚未实现，因此这里明确返回占位值，而不是伪造监控数据。
	return SystemStatus{
		ProblemCount:    problemCount,
		SubmissionCount: submissionCount,
		PendingJudges:   pendingCount,
		JudgerStatus:    "stubbed",
		ConcurrentJobs:  0,
	}, nil
}
