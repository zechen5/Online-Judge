// Package repository isolates database access and persistence behavior.
// repository 包隔离数据库访问与持久化行为。
package repositories

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"online-judge-backend/internal/config"
	"online-judge-backend/internal/models"
)

func NewDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	// Switch database driver by configuration so tests can use SQLite while
	// deployment can still target MySQL.
	// 根据配置切换数据库驱动，这样测试可用 SQLite，部署仍可使用 MySQL。
	switch cfg.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	case "sqlite", "":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
	if err != nil {
		return nil, err
	}

	// Auto-migrate all currently supported domain models.
	// 自动迁移当前支持的全部领域模型。
	if err := db.AutoMigrate(&models.User{}, &models.Problem{}, &models.TestCase{}, &models.Submission{}); err != nil {
		return nil, err
	}

	// Ensure a default admin exists so the admin middleware has a usable account
	// immediately after first boot.
	// 确保默认管理员账号存在，这样管理员中间件在首次启动后就有可用账号。
	if err := ensureDefaultAdmin(db, cfg); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureDefaultAdmin(db *gorm.DB, _ config.DatabaseConfig) error {
	username := "admin"
	password := "admin123456"

	if value := getenv("OJ_ADMIN_USERNAME", ""); value != "" {
		username = value
	}
	if value := getenv("OJ_ADMIN_PASSWORD", ""); value != "" {
		password = value
	}

	var user models.User
	err := db.Where("username = ?", username).First(&user).Error
	if err == nil {
		if user.Role != models.UserRoleAdmin {
			user.Role = models.UserRoleAdmin
			return db.Save(&user).Error
		}
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.Create(&models.User{
		Username: username,
		Password: string(hash),
		Role:     models.UserRoleAdmin,
	}).Error
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
