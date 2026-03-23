// Package config loads runtime configuration for the application.
// config 包负责加载应用运行时配置。
package config

import (
	"fmt"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Judge    JudgeConfig
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

// JudgeConfig groups asynchronous judging runtime settings.
// JudgeConfig 聚合异步判题运行时配置。
type JudgeConfig struct {
	WorkerCount   int
	QueueCapacity int
	Executor      string
	Docker        DockerJudgeConfig
}

// DockerJudgeConfig stores container runtime settings for the Docker evaluator.
// DockerJudgeConfig 存储 Docker 判题器的容器运行配置。
type DockerJudgeConfig struct {
	CPPImage       string
	JavaImage      string
	PythonImage    string
	CompileMemory  string
	RuntimeMemory  string
	CPULimit       string
	WorkspacePath  string
	CompileTimeout int
}

// Load reads config from environment variables and fills sensible defaults.
// Load 从环境变量读取配置，并填充合理的默认值。
func Load() Config {
	driver := getenv("OJ_DB_DRIVER", "sqlite")
	dsn := getenv("OJ_DB_DSN", "app.db")
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
		Judge: JudgeConfig{
			WorkerCount:   getenvInt("OJ_JUDGE_WORKERS", 2),
			QueueCapacity: getenvInt("OJ_JUDGE_QUEUE_CAPACITY", 64),
			Executor:      getenv("OJ_JUDGE_EXECUTOR", "auto"),
			Docker: DockerJudgeConfig{
				CPPImage:       getenv("OJ_DOCKER_CPP_IMAGE", "algojudge/judge-cpp:latest"),
				JavaImage:      getenv("OJ_DOCKER_JAVA_IMAGE", "algojudge/judge-java:latest"),
				PythonImage:    getenv("OJ_DOCKER_PYTHON_IMAGE", "algojudge/judge-python:latest"),
				CompileMemory:  getenv("OJ_DOCKER_COMPILE_MEMORY", "768m"),
				RuntimeMemory:  getenv("OJ_DOCKER_RUNTIME_MEMORY", "512m"),
				CPULimit:       getenv("OJ_DOCKER_CPU_LIMIT", "1.0"),
				WorkspacePath:  getenv("OJ_DOCKER_WORKSPACE", "/workspace"),
				CompileTimeout: getenvInt("OJ_DOCKER_COMPILE_TIMEOUT_SECONDS", 20),
			},
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

// getenvInt reads an integer environment variable with a fallback.
// getenvInt 读取整数环境变量，并在缺失或非法时回退到默认值。
func getenvInt(key string, fallback int) int {
	value := getenv(key, "")
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
