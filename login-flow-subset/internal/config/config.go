package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	DSN string
}

type AuthConfig struct {
	JWTSecret            string
	DefaultAdminUsername string
	DefaultAdminPassword string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getenv("OJ_LOGIN_SUBSET_PORT", "8090"),
		},
		Database: DatabaseConfig{
			DSN: getenv("OJ_LOGIN_SUBSET_DSN", "login-flow-subset.db"),
		},
		Auth: AuthConfig{
			JWTSecret:            getenv("OJ_LOGIN_SUBSET_JWT_SECRET", "login-subset-secret"),
			DefaultAdminUsername: getenv("OJ_LOGIN_SUBSET_ADMIN_USERNAME", "admin"),
			DefaultAdminPassword: getenv("OJ_LOGIN_SUBSET_ADMIN_PASSWORD", "admin123456"),
		},
	}
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
