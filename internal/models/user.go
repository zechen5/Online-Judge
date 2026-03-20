// Package models contains GORM entities shared across repositories and services.
// models 包存放仓储层与服务层共享的 GORM 实体定义。
package models

import "time"

const (
	// UserRoleStudent is the default role for normal users.
	// UserRoleStudent 是普通用户的默认角色。
	UserRoleStudent = 0
	// UserRoleAdmin is the role for administrators.
	// UserRoleAdmin 是管理员角色。
	UserRoleAdmin = 1
)

type User struct {
	// ID is the primary key for the user table.
	// ID 是用户表的主键。
	ID uint `gorm:"primaryKey" json:"id"`
	// Username is unique and used in leaderboard display.
	// Username 必须唯一，并用于排行榜展示。
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	// Password stores a hash placeholder for now, never plain text in API output.
	// Password 当前存储哈希占位值，且永远不会直接通过 API 输出。
	Password string `gorm:"size:255;not null" json:"-"`
	// Role distinguishes normal users and admins.
	// Role 用于区分普通用户和管理员。
	Role int `gorm:"default:0;not null" json:"role"`
	// CreatedAt and UpdatedAt are maintained by GORM automatically.
	// CreatedAt 和 UpdatedAt 由 GORM 自动维护。
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
