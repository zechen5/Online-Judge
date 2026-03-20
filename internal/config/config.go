// Package config loads runtime configuration for the application.
// config 包负责加载应用运行时配置。
package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

// ServerConfig groups HTTP server settings.
// ServerConfig 聚合 HTTP 服务相关配置。
type ServerConfig struct {
	Port string
}

// DatabaseConfig groups DB driver and DSN settings.
// DatabaseConfig 聚合数据库驱动和 DSN 配置。
type DatabaseConfig struct {
	Driver string
	DSN    string
}

// AuthConfig groups token and bootstrap-account settings.
// AuthConfig 聚合令牌和初始化账号相关配置。
type AuthConfig struct {
	JWTSecret            string
	DefaultAdminUsername string
	DefaultAdminPassword string
}

// Load reads config from environment variables and fills sensible defaults.
// Load 从环境变量读取配置，并填充合理的默认值。
func Load() Config {
	driver := getenv("OJ_DB_DRIVER", "sqlite")
	dsn := getenv("OJ_DB_DSN", "app.db")
	// When MySQL is selected without a custom DSN, provide a local default.
	// 当选择 MySQL 但未提供自定义 DSN 时，补一个本地默认值。
	if driver == "mysql" && dsn == "app.db" {
		dsn = "root:password@tcp(127.0.0.1:3306)/online_judge?charset=utf8mb4&parseTime=True&loc=Local"
	}

	return Config{
		Server: ServerConfig{Port: getenv("OJ_PORT", "8080")},
		Database: DatabaseConfig{
			Driver: driver,
			DSN:    dsn,
		},
		Auth: AuthConfig{
			JWTSecret:            getenv("OJ_JWT_SECRET", "dev-secret-change-me"),
			DefaultAdminUsername: getenv("OJ_ADMIN_USERNAME", "admin"),
			DefaultAdminPassword: getenv("OJ_ADMIN_PASSWORD", "admin123456"),
		},
	}
}

// getenv wraps os.LookupEnv so callers can request a fallback inline.
// getenv 封装 os.LookupEnv，便于调用方就地提供回退值。
func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
