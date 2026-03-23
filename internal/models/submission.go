// Package models contains GORM entities shared across repositories and services.
// models 包存放仓储层与服务层共享的 GORM 实体定义。
package models

import "time"

const (
	// SubmissionStatusPending means the submission is queued and not started yet.
	// SubmissionStatusPending 表示提交已入队，但尚未开始评测。
	SubmissionStatusPending = "Pending"
	// SubmissionStatusJudging means the worker is actively evaluating the submission.
	// SubmissionStatusJudging 表示工作线程正在实际评测该提交。
	SubmissionStatusJudging = "Judging"
	// SubmissionStatusAC means the submission passed every testcase.
	// SubmissionStatusAC 表示提交通过了全部测试点。
	SubmissionStatusAC = "AC"
	// SubmissionStatusWA means the output differs from the expected answer.
	// SubmissionStatusWA 表示输出与标准答案不一致。
	SubmissionStatusWA = "WA"
	// SubmissionStatusTLE means the execution exceeded the time limit.
	// SubmissionStatusTLE 表示执行超出了时间限制。
	SubmissionStatusTLE = "TLE"
	// SubmissionStatusRE means the program exited abnormally during execution.
	// SubmissionStatusRE 表示程序在执行期间异常退出。
	SubmissionStatusRE = "RE"
	// SubmissionStatusCE means compilation failed before execution started.
	// SubmissionStatusCE 表示在执行前编译失败。
	SubmissionStatusCE = "CE"
)

type Submission struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `gorm:"index;not null" json:"user_id"`
	ProblemID uint   `gorm:"index;not null" json:"problem_id"`
	Language  string `gorm:"size:20;not null" json:"language"`
	Code      string `gorm:"type:text;not null" json:"code"`
	Status    string `gorm:"size:20;default:'Pending';not null" json:"status"`
	Runtime   int    `gorm:"not null;default:0" json:"runtime"`
	Memory    int    `gorm:"not null;default:0" json:"memory"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`

	User    User    `json:"user,omitempty"`
	Problem Problem `json:"problem,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
