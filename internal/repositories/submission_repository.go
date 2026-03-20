// Package repository isolates database access and persistence behavior.
// repository 包隔离数据库访问与持久化行为。
package repositories

import (
	"online-judge-backend/internal/models"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates the submission persistence gateway.
// NewSubmissionRepository 创建提交记录的数据访问网关。
func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// Create inserts a new submission row.
// Create 插入一条新的提交记录。
func (r *SubmissionRepository) Create(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

// Update writes back judge results onto an existing submission.
// Update 把判题结果回写到已有提交记录中。
func (r *SubmissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

// GetByID returns one submission with its user relation preloaded.
// GetByID 返回一条提交记录，并预加载用户信息。
func (r *SubmissionRepository) GetByID(id uint) (*models.Submission, error) {
	var submission models.Submission
	if err := r.db.Preload("User").First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

// TopRank fetches the fastest accepted submissions for a problem.
// TopRank 获取指定题目用时最短的通过提交。
func (r *SubmissionRepository) TopRank(problemID uint, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	// Leaderboard uses AC-only results ordered by runtime then submission time.
	// 排行榜只统计 AC 结果，并按运行时间、提交时间排序。
	err := r.db.Preload("User").
		Where("problem_id = ? AND status = ?", problemID, models.SubmissionStatusAC).
		Order("runtime asc, created_at asc, id asc").
		Limit(limit).
		Find(&submissions).Error
	return submissions, err
}

// CountPending reports how many submissions are still waiting on a judge result.
// CountPending 返回仍在等待判题结果的提交数量。
func (r *SubmissionRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&models.Submission{}).Where("status = ?", models.SubmissionStatusPending).Count(&count).Error
	return count, err
}

// CountAll reports all submissions regardless of status.
// CountAll 返回所有状态下的提交总数。
func (r *SubmissionRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.Submission{}).Count(&count).Error
	return count, err
}
