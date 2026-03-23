// Package repositories isolates database access and persistence behavior.
// repositories 包隔离数据库访问与持久化行为。
package repositories

import (
	"online-judge-backend/internal/models"

	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

func (r *SubmissionRepository) GetByID(id uint) (*models.Submission, error) {
	var submission models.Submission
	if err := r.db.Preload("User").Preload("Problem").First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) ListByUser(userID uint, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Preload("Problem").
		Where("user_id = ?", userID).
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&submissions).Error
	return submissions, err
}

func (r *SubmissionRepository) TopRank(problemID uint, limit int) ([]models.Submission, error) {
	var submissions []models.Submission
	err := r.db.Preload("User").
		Where("problem_id = ? AND status = ?", problemID, models.SubmissionStatusAC).
		Order("runtime asc, created_at asc, id asc").
		Limit(limit).
		Find(&submissions).Error
	return submissions, err
}

func (r *SubmissionRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&models.Submission{}).
		Where("status IN ?", []string{models.SubmissionStatusPending, models.SubmissionStatusJudging}).
		Count(&count).Error
	return count, err
}

func (r *SubmissionRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.Submission{}).Count(&count).Error
	return count, err
}
