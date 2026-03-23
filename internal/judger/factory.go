// Package judger contains pluggable judge implementations.
// judger 包存放可插拔的判题实现。
package judger

import (
	"fmt"
	"strings"

	"online-judge-backend/internal/config"
	"online-judge-backend/internal/services"
)

// BuildEvaluator selects a configured evaluator implementation and reports its mode.
// BuildEvaluator 负责按配置选择判题实现，并返回最终使用的模式名称。
func BuildEvaluator(cfg config.JudgeConfig) (services.Evaluator, string, error) {
	executor := strings.ToLower(strings.TrimSpace(cfg.Executor))
	if executor == "" {
		executor = "auto"
	}

	switch executor {
	case "local":
		return NewLocalEvaluator(), "async-local", nil
	case "docker":
		evaluator, err := NewDockerEvaluator(cfg.Docker)
		if err != nil {
			return nil, "", err
		}
		if err := evaluator.Ping(); err != nil {
			return nil, "", err
		}
		return evaluator, "async-docker", nil
	case "auto":
		evaluator, err := NewDockerEvaluator(cfg.Docker)
		if err == nil && evaluator.Ping() == nil {
			return evaluator, "async-docker", nil
		}
		return NewLocalEvaluator(), "async-local", nil
	default:
		return nil, "", fmt.Errorf("unsupported judge executor %q", cfg.Executor)
	}
}
