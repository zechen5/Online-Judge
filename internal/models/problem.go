// Package models contains GORM entities shared across repositories and services.
// models 包存放仓储层与服务层共享的 GORM 实体定义。
package models

import "time"

const (
	// ProblemStatusPending means the problem still awaits review.
	// ProblemStatusPending 表示题目仍处于待审核状态。
	ProblemStatusPending = 0
	// ProblemStatusPublished means the problem is visible for browsing and submission.
	// ProblemStatusPublished 表示题目已发布，可被浏览和提交。
	ProblemStatusPublished = 1
	// ProblemStatusHidden means the problem is intentionally hidden from public views.
	// ProblemStatusHidden 表示题目已存在，但被刻意隐藏，不对公众展示。
	ProblemStatusHidden = 2
)

type Problem struct {
	// ID is the primary key for problem records.
	// ID 是题目记录的主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// Title is the short visible name shown in lists and detail pages.
	// Title 是列表页和详情页展示的题目名称。
	Title string `gorm:"size:100;not null" json:"title"`
	// Description stores the full statement, usually in Markdown.
	// Description 存储完整题面，通常使用 Markdown。
	Description string `gorm:"type:text" json:"description"`
	// TimeLimit is expressed in milliseconds.
	// TimeLimit 以毫秒为单位。
	TimeLimit int `gorm:"not null" json:"time_limit"`
	// MemoryLimit is the submission memory ceiling.
	// MemoryLimit 是提交运行时的内存上限。
	MemoryLimit int `gorm:"not null" json:"memory_limit"`
	// Status controls whether the problem is public, pending review, or hidden.
	// Status 控制题目是公开、待审核还是隐藏。
	Status int `gorm:"default:0;not null" json:"status"`
	// CreatorID links the problem back to its uploader.
	// CreatorID 把题目关联回上传者。
	CreatorID uint `gorm:"index;not null" json:"creator_id"`
	// TestCases holds both sample and hidden cases.
	// TestCases 同时包含样例和隐藏测试用例。
	TestCases []TestCase `json:"test_cases,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TestCase struct {
	// ID is the primary key for each testcase row.
	// ID 是每条测试用例记录的主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// ProblemID points to the owning problem.
	// ProblemID 指向所属题目。
	ProblemID uint `gorm:"index;not null" json:"problem_id"`
	// Input and Output store the testcase pair.
	// Input 和 Output 存储测试用例输入输出对。
	Input  string `gorm:"type:text;not null" json:"input"`
	Output string `gorm:"type:text;not null" json:"output"`
	// IsSample marks cases that can be exposed in problem detail responses.
	// IsSample 标记该用例是否可在题目详情中作为样例公开。
	IsSample bool `gorm:"default:false;not null" json:"is_sample"`
	// SortOrder preserves the author-defined testcase order.
	// SortOrder 保留出题人设定的测试用例顺序。
	SortOrder int       `gorm:"default:0;not null" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
