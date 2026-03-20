// Package repository isolates database access and persistence behavior.
// repository 包隔离数据库访问与持久化行为。
package repositories

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"online-judge-backend/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates the user persistence gateway.
// NewUserRepository 创建用户数据访问网关。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row.
// Create 插入一条新的用户记录。
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Save updates an existing user row.
// Save 更新一条已有的用户记录。
func (r *UserRepository) Save(user *models.User) error {
	return r.db.Save(user).Error
}

// GetByID returns a user by primary key.
// GetByID 按主键返回用户。
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername returns a user by unique username.
// GetByUsername 按唯一用户名返回用户。
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// EnsureUser guarantees the referenced user exists, creating a placeholder
// record when the caller only provides a user ID.
// EnsureUser 保证被引用的用户存在；当调用方只传 user ID 时，会创建占位用户。
func (r *UserRepository) EnsureUser(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// This scaffold does not implement registration yet, so a deterministic
	// placeholder username keeps submission and leaderboard flows usable.
	// 当前骨架尚未实现注册，因此使用确定性的占位用户名保证提交和排行榜可用。
	user = models.User{
		ID:       id,
		Username: fmt.Sprintf("user_%d", id),
		Password: "placeholder",
		Role:     models.UserRoleStudent,
	}
	if err := r.db.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
