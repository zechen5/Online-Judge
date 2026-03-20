// Package models contains GORM entities shared across repositories and services.
// models 包存放仓储层与服务层共享的 GORM 实体定义。
package models

import "time"

const (
	// SubmissionStatusPending marks work that has not finished judging yet.
	// SubmissionStatusPending 表示该提交尚未完成判题。
	SubmissionStatusPending = "Pending"
	// SubmissionStatusAC means the submission passed all cases.
	// SubmissionStatusAC 表示提交通过全部测试用例。
	SubmissionStatusAC = "AC"
	// SubmissionStatusWA means wrong answer.
	// SubmissionStatusWA 表示答案错误。
	SubmissionStatusWA = "WA"
	// SubmissionStatusTLE means time limit exceeded.
	// SubmissionStatusTLE 表示超出时间限制。
	SubmissionStatusTLE = "TLE"
	// SubmissionStatusRE means runtime error.
	// SubmissionStatusRE 表示运行时错误。
	SubmissionStatusRE = "RE"
	// SubmissionStatusCE means compile error.
	// SubmissionStatusCE 表示编译错误。
	SubmissionStatusCE = "CE"
)

type Submission struct {
	// ID is the primary key for submission records.
	// ID 是提交记录的主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// UserID and ProblemID form the submission ownership relation.
	// UserID 和 ProblemID 共同描述这次提交属于谁、对应哪道题。
	UserID    uint `gorm:"index;not null" json:"user_id"`
	ProblemID uint `gorm:"index;not null" json:"problem_id"`
	// Language and Code capture the submitted source code payload.
	// Language 和 Code 记录用户提交的语言与源码内容。
	Language string `gorm:"size:20;not null" json:"language"`
	Code     string `gorm:"type:text;not null" json:"code"`
	// Status, Runtime, Memory, and ErrorMsg are judge outputs.
	// Status、Runtime、Memory 和 ErrorMsg 都是判题输出字段。
	Status   string `gorm:"size:20;default:'Pending';not null" json:"status"`
	Runtime  int    `gorm:"not null;default:0" json:"runtime"`
	Memory   int    `gorm:"not null;default:0" json:"memory"`
	ErrorMsg string `gorm:"type:text" json:"error_msg"`
	// User is preloaded when leaderboard or detail output needs usernames.
	// 当排行榜或详情需要用户名时，会预加载 User。
	User      User      `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
