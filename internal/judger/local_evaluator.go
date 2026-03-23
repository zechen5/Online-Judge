// Package judger contains pluggable judge implementations.
// judger 包存放可插拔的判题实现。
package judger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"online-judge-backend/internal/models"
	"online-judge-backend/internal/services"
)

// LocalEvaluator compiles and runs submissions directly on the host machine.
// LocalEvaluator 会在宿主机上直接编译并运行提交代码。
type LocalEvaluator struct {
	drivers        map[string]languageDriver
	compileTimeout time.Duration
}

type languageDriver interface {
	Compile(ctx context.Context, workDir string, code string) (*compiledProgram, error)
}

type compiledProgram struct {
	run func(ctx context.Context, input string) (executionResult, error)
}

type executionResult struct {
	Stdout   string
	Stderr   string
	Runtime  int
	Memory   int
	ExitCode int
	TimedOut bool
}

type cppDriver struct{}
type javaDriver struct{}
type pythonDriver struct{}

// NewLocalEvaluator creates a local multi-language evaluator.
// NewLocalEvaluator 创建一个本地多语言评测器。
func NewLocalEvaluator() *LocalEvaluator {
	return &LocalEvaluator{
		drivers: map[string]languageDriver{
			"cpp":    cppDriver{},
			"java":   javaDriver{},
			"python": pythonDriver{},
		},
		compileTimeout: 15 * time.Second,
	}
}

// Evaluate compiles the submission and executes it against every testcase.
// Evaluate 会先编译提交，再对每个测试点执行程序。
func (e *LocalEvaluator) Evaluate(ctx context.Context, problem *models.Problem, submission *models.Submission, testCases []models.TestCase) (services.JudgeResult, error) {
	if len(testCases) == 0 {
		return services.JudgeResult{
			Status:   models.SubmissionStatusRE,
			ErrorMsg: "no testcases configured",
		}, nil
	}

	driver, ok := e.drivers[strings.ToLower(strings.TrimSpace(submission.Language))]
	if !ok {
		return services.JudgeResult{
			Status:   models.SubmissionStatusCE,
			ErrorMsg: "unsupported language",
		}, nil
	}

	workDir, err := os.MkdirTemp("", "algojudge-*")
	if err != nil {
		return services.JudgeResult{}, err
	}
	defer os.RemoveAll(workDir)

	compileCtx, cancelCompile := context.WithTimeout(ctx, e.compileTimeout)
	defer cancelCompile()

	program, err := driver.Compile(compileCtx, workDir, submission.Code)
	if err != nil {
		return services.JudgeResult{
			Status:   models.SubmissionStatusCE,
			ErrorMsg: trimOutput(err.Error()),
		}, nil
	}

	maxRuntime := 0
	maxMemory := 0
	for _, tc := range testCases {
		runCtx, cancelRun := context.WithTimeout(ctx, time.Duration(problem.TimeLimit)*time.Millisecond)
		result, runErr := program.run(runCtx, tc.Input)
		cancelRun()

		if runErr != nil {
			if result.TimedOut {
				return services.JudgeResult{
					Status:   models.SubmissionStatusTLE,
					Runtime:  max(maxRuntime, result.Runtime),
					Memory:   maxMemory,
					ErrorMsg: "time limit exceeded",
				}, nil
			}
			return services.JudgeResult{
				Status:   models.SubmissionStatusRE,
				Runtime:  max(maxRuntime, result.Runtime),
				Memory:   max(maxMemory, result.Memory),
				ErrorMsg: trimOutput(runErr.Error()),
			}, nil
		}

		if result.TimedOut {
			return services.JudgeResult{
				Status:   models.SubmissionStatusTLE,
				Runtime:  max(maxRuntime, result.Runtime),
				Memory:   maxMemory,
				ErrorMsg: "time limit exceeded",
			}, nil
		}
		if result.ExitCode != 0 {
			return services.JudgeResult{
				Status:   models.SubmissionStatusRE,
				Runtime:  max(maxRuntime, result.Runtime),
				Memory:   max(maxMemory, result.Memory),
				ErrorMsg: trimOutput(result.Stderr),
			}, nil
		}

		maxRuntime = max(maxRuntime, result.Runtime)
		maxMemory = max(maxMemory, result.Memory)
		if !outputsMatch(result.Stdout, tc.Output) {
			return services.JudgeResult{
				Status:   models.SubmissionStatusWA,
				Runtime:  maxRuntime,
				Memory:   maxMemory,
				ErrorMsg: "",
			}, nil
		}
	}

	return services.JudgeResult{
		Status:  models.SubmissionStatusAC,
		Runtime: maxRuntime,
		Memory:  maxMemory,
	}, nil
}

func (cppDriver) Compile(ctx context.Context, workDir string, code string) (*compiledProgram, error) {
	sourcePath := filepath.Join(workDir, "main.cpp")
	exePath := filepath.Join(workDir, "main.exe")
	if err := os.WriteFile(sourcePath, []byte(code), 0o644); err != nil {
		return nil, err
	}

	output, err := runCombined(ctx, workDir, "g++", "-O2", "-std=c++17", "-o", exePath, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("compile failed: %s", trimOutput(output))
	}

	return &compiledProgram{
		run: func(ctx context.Context, input string) (executionResult, error) {
			return runExecutable(ctx, workDir, input, exePath)
		},
	}, nil
}

func (javaDriver) Compile(ctx context.Context, workDir string, code string) (*compiledProgram, error) {
	sourcePath := filepath.Join(workDir, "Main.java")
	if err := os.WriteFile(sourcePath, []byte(code), 0o644); err != nil {
		return nil, err
	}

	output, err := runCombined(ctx, workDir, "javac", sourcePath)
	if err != nil {
		return nil, fmt.Errorf("compile failed: %s", trimOutput(output))
	}

	return &compiledProgram{
		run: func(ctx context.Context, input string) (executionResult, error) {
			return runExecutable(ctx, workDir, input, "java", "-cp", workDir, "Main")
		},
	}, nil
}

func (pythonDriver) Compile(ctx context.Context, workDir string, code string) (*compiledProgram, error) {
	sourcePath := filepath.Join(workDir, "main.py")
	if err := os.WriteFile(sourcePath, []byte(code), 0o644); err != nil {
		return nil, err
	}

	output, err := runCombined(ctx, workDir, "python", "-m", "py_compile", sourcePath)
	if err != nil {
		return nil, fmt.Errorf("compile failed: %s", trimOutput(output))
	}

	return &compiledProgram{
		run: func(ctx context.Context, input string) (executionResult, error) {
			return runExecutable(ctx, workDir, input, "python", sourcePath)
		},
	}, nil
}

func runCombined(ctx context.Context, workDir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("command timed out")
	}
	return string(output), err
}

func runExecutable(ctx context.Context, workDir string, input string, name string, args ...string) (executionResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	cmd.Stdin = strings.NewReader(input)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startedAt := time.Now()
	err := cmd.Run()
	runtimeMS := int(time.Since(startedAt).Milliseconds())

	result := executionResult{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Runtime: runtimeMS,
		Memory:  0,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result, ctx.Err()
	}

	if err == nil {
		return result, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	return result, err
}

func outputsMatch(actual string, expected string) bool {
	return normalizeOutput(actual) == normalizeOutput(expected)
}

func normalizeOutput(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.TrimSpace(value)
}

func trimOutput(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 800 {
		return value[:800] + "..."
	}
	return value
}

func max(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
