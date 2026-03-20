// Package repository isolates database access and persistence behavior.
// repository 包隔离数据库访问与持久化行为。
package repositories

import (
	"gorm.io/gorm"

	"online-judge-backend/internal/models"
)

type ProblemRepository struct {
	db *gorm.DB
}

// NewProblemRepository creates the problem persistence gateway.
// NewProblemRepository 创建题目数据访问网关。
func NewProblemRepository(db *gorm.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

// Create inserts a problem and its associated testcases.
// Create 会写入题目及其关联的测试用例。
func (r *ProblemRepository) Create(problem *models.Problem) error {
	return r.db.Create(problem).Error
}

// Update persists the parent problem and fully replaces its testcase set.
// Update 会保存题目本身，并完整替换其测试用例集合。
func (r *ProblemRepository) Update(problem *models.Problem) error {
	// Replace keeps testcase associations aligned with the request payload.
	// Replace 让测试用例关联与请求载荷保持一致。
	if err := r.db.Model(problem).Association("TestCases").Replace(problem.TestCases); err != nil {
		return err
	}
	// Omit prevents GORM from trying to save associations twice.
	// Omit 用于避免 GORM 对关联对象重复保存。
	return r.db.Omit("TestCases").Save(problem).Error
}

// GetByID loads a problem together with ordered testcases.
// GetByID 会连同排序后的测试用例一起加载题目。
func (r *ProblemRepository) GetByID(id uint) (*models.Problem, error) {
	var problem models.Problem
	// Keep testcase order stable for sample rendering and future judging.
	// 保持测试用例顺序稳定，便于样例展示和后续判题。
	if err := r.db.Preload("TestCases", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order asc, id asc")
	}).First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

// ListPublished returns public problems only, with paging metadata.
// ListPublished 仅返回已发布题目，并支持分页。
func (r *ProblemRepository) ListPublished(offset, limit int) ([]models.Problem, int64, error) {
	var (
		problems []models.Problem
		total    int64
	)

	// Count first so the API can return total rows for pagination.
	// 先统计总数，这样 API 能返回分页总量。
	query := r.db.Model(&models.Problem{}).Where("status = ?", models.ProblemStatusPublished)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id asc").Offset(offset).Limit(limit).Find(&problems).Error; err != nil {
		return nil, 0, err
	}
	return problems, total, nil
}

// Count returns the total number of stored problems.
// Count 返回当前存储的题目总数。
func (r *ProblemRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Problem{}).Count(&count).Error
	return count, err
}
