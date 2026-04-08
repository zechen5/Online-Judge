package repositories

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"online-judge-backend/login-flow-subset/internal/config"
	"online-judge-backend/login-flow-subset/internal/models"
)

func NewDatabase(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, err
	}

	if err := ensureDefaultAdmin(db, cfg.Auth); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureDefaultAdmin(db *gorm.DB, cfg config.AuthConfig) error {
	var user models.User
	err := db.Where("username = ?", cfg.DefaultAdminUsername).First(&user).Error
	if err == nil {
		if user.Role != models.UserRoleAdmin {
			user.Role = models.UserRoleAdmin
			return db.Save(&user).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.Create(&models.User{
		Username: cfg.DefaultAdminUsername,
		Password: string(hash),
		Role:     models.UserRoleAdmin,
	}).Error
}
