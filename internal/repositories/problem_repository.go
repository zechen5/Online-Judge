// Package repositories isolates database access and persistence behavior.
// repositories 包隔离数据库访问与持久化行为。
package repositories

import (
	"fmt"

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
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("TestCases").Save(problem).Error; err != nil {
			return err
		}

		if err := tx.Where("problem_id = ?", problem.ID).Delete(&models.TestCase{}).Error; err != nil {
			return err
		}

		for index := range problem.TestCases {
			problem.TestCases[index].ID = 0
			problem.TestCases[index].ProblemID = problem.ID
		}
		if len(problem.TestCases) == 0 {
			return nil
		}
		return tx.Create(&problem.TestCases).Error
	})
}

// Delete removes the target problem and shifts following IDs left by one.
// Delete 会删除目标题目，并把后续题目编号整体前移一位。
func (r *ProblemRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("problem_id = ?", id).Delete(&models.TestCase{}).Error; err != nil {
			return err
		}
		if err := tx.Where("problem_id = ?", id).Delete(&models.Submission{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Problem{}, id).Error; err != nil {
			return err
		}

		// Shift dependent foreign keys first so later list/detail requests still
		// point at the same logical problems after the primary key compaction.
		// 先平移依赖表中的外键，这样主键压缩后，查询仍然指向原来的逻辑题目。
		if err := tx.Exec("UPDATE test_cases SET problem_id = problem_id - 1 WHERE problem_id > ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE submissions SET problem_id = problem_id - 1 WHERE problem_id > ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE problems SET id = id - 1 WHERE id > ?", id).Error; err != nil {
			return err
		}

		return r.reseedProblemIDs(tx)
	})
}

// GetByID loads a problem together with ordered testcases.
// GetByID 会连同排好序的测试用例一起加载题目。
func (r *ProblemRepository) GetByID(id uint) (*models.Problem, error) {
	var problem models.Problem
	if err := r.db.Preload("TestCases", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order asc, id asc")
	}).First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

// GetPublishedByID loads one public problem and its ordered testcase set.
// GetPublishedByID 会加载一个公开题目及其排好序的测试用例。
func (r *ProblemRepository) GetPublishedByID(id uint) (*models.Problem, error) {
	var problem models.Problem
	if err := r.db.Preload("TestCases", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order asc, id asc")
	}).Where("status = ?", models.ProblemStatusPublished).First(&problem, id).Error; err != nil {
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

	query := r.db.Model(&models.Problem{}).Where("status = ?", models.ProblemStatusPublished)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id asc").Offset(offset).Limit(limit).Find(&problems).Error; err != nil {
		return nil, 0, err
	}
	return problems, total, nil
}

// ListAll returns all problems for admin review, with paging metadata.
// ListAll 返回管理员审核所需的全量题目，并附带分页统计。
func (r *ProblemRepository) ListAll(offset, limit int) ([]models.Problem, int64, error) {
	var (
		problems []models.Problem
		total    int64
	)

	query := r.db.Model(&models.Problem{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("status asc, updated_at desc, id desc").Offset(offset).Limit(limit).Find(&problems).Error; err != nil {
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

func (r *ProblemRepository) reseedProblemIDs(tx *gorm.DB) error {
	var nextID int64 = 1
	if err := tx.Model(&models.Problem{}).Select("COALESCE(MAX(id), 0) + 1").Scan(&nextID).Error; err != nil {
		return err
	}

	switch tx.Dialector.Name() {
	case "sqlite":
		// Deleting the sequence row lets SQLite recalculate from MAX(id) on the
		// next insert, which keeps numbering contiguous after manual ID shifts.
		// 删除序列表后，SQLite 会在下一次插入时基于 MAX(id) 重新计算，可保证手动平移后的编号连续。
		_ = tx.Exec("DELETE FROM sqlite_sequence WHERE name = ?", "problems").Error
	case "mysql":
		if err := tx.Exec(fmt.Sprintf("ALTER TABLE problems AUTO_INCREMENT = %d", nextID)).Error; err != nil {
			return err
		}
	}

	return nil
}
