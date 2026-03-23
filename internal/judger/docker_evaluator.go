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

	"online-judge-backend/internal/config"
	"online-judge-backend/internal/models"
	"online-judge-backend/internal/services"
)

// DockerEvaluator compiles and runs submissions inside disposable Docker containers.
// DockerEvaluator 会在一次性 Docker 容器中编译并运行提交代码。
type DockerEvaluator struct {
	images         map[string]string
	workspacePath  string
	compileTimeout time.Duration
	compileMemory  string
	runtimeMemory  string
	cpuLimit       string
}

type dockerProgram struct {
	run func(ctx context.Context, input string) (executionResult, error)
}

// NewDockerEvaluator constructs the Docker-backed evaluator.
// NewDockerEvaluator 构造基于 Docker 的判题实现。
func NewDockerEvaluator(cfg config.DockerJudgeConfig) (*DockerEvaluator, error) {
	if cfg.WorkspacePath == "" {
		cfg.WorkspacePath = "/workspace"
	}
	if cfg.CompileTimeout <= 0 {
		cfg.CompileTimeout = 20
	}
	return &DockerEvaluator{
		images: map[string]string{
			"cpp":    cfg.CPPImage,
			"java":   cfg.JavaImage,
			"python": cfg.PythonImage,
		},
		workspacePath:  cfg.WorkspacePath,
		compileTimeout: time.Duration(cfg.CompileTimeout) * time.Second,
		compileMemory:  cfg.CompileMemory,
		runtimeMemory:  cfg.RuntimeMemory,
		cpuLimit:       cfg.CPULimit,
	}, nil
}

// Ping verifies that the Docker daemon is reachable before the evaluator is used.
// Ping 用于在正式评测前确认 Docker daemon 可达。
func (e *DockerEvaluator) Ping() error {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker unavailable: %s", trimOutput(string(output)))
	}
	return nil
}

// Evaluate compiles the submission and executes it against every testcase.
// Evaluate 会先编译提交，再对每个测试点执行程序。
func (e *DockerEvaluator) Evaluate(ctx context.Context, problem *models.Problem, submission *models.Submission, testCases []models.TestCase) (services.JudgeResult, error) {
	if len(testCases) == 0 {
		return services.JudgeResult{
			Status:   models.SubmissionStatusRE,
			ErrorMsg: "no testcases configured",
		}, nil
	}

	language := strings.ToLower(strings.TrimSpace(submission.Language))
	image, ok := e.images[language]
	if !ok || image == "" {
		return services.JudgeResult{
			Status:   models.SubmissionStatusCE,
			ErrorMsg: "unsupported language",
		}, nil
	}

	workDir, err := os.MkdirTemp("", "algojudge-docker-*")
	if err != nil {
		return services.JudgeResult{}, err
	}
	defer os.RemoveAll(workDir)

	program, err := e.prepareProgram(ctx, image, language, workDir, submission.Code)
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
				Status:  models.SubmissionStatusWA,
				Runtime: maxRuntime,
				Memory:  maxMemory,
			}, nil
		}
	}

	return services.JudgeResult{
		Status:  models.SubmissionStatusAC,
		Runtime: maxRuntime,
		Memory:  maxMemory,
	}, nil
}

func (e *DockerEvaluator) prepareProgram(ctx context.Context, image string, language string, workDir string, code string) (*dockerProgram, error) {
	var (
		sourceName string
		compileCmd []string
		runCmd     []string
	)

	switch language {
	case "cpp":
		sourceName = "main.cpp"
		compileCmd = []string{"sh", "-lc", fmt.Sprintf("g++ -O2 -std=c++17 -o %s/main %s/main.cpp", e.workspacePath, e.workspacePath)}
		runCmd = []string{fmt.Sprintf("%s/main", e.workspacePath)}
	case "java":
		sourceName = "Main.java"
		compileCmd = []string{"javac", filepath.ToSlash(filepath.Join(e.workspacePath, sourceName))}
		runCmd = []string{"java", "-cp", e.workspacePath, "Main"}
	case "python":
		sourceName = "main.py"
		compileCmd = []string{"python", "-m", "py_compile", filepath.ToSlash(filepath.Join(e.workspacePath, sourceName))}
		runCmd = []string{"python", filepath.ToSlash(filepath.Join(e.workspacePath, sourceName))}
	default:
		return nil, fmt.Errorf("unsupported language %q", language)
	}

	sourcePath := filepath.Join(workDir, sourceName)
	if err := os.WriteFile(sourcePath, []byte(code), 0o644); err != nil {
		return nil, err
	}

	compileCtx, cancelCompile := context.WithTimeout(ctx, e.compileTimeout)
	defer cancelCompile()

	if output, result, err := e.runContainer(compileCtx, image, workDir, "", e.compileMemory, compileCmd...); err != nil {
		if result.TimedOut {
			return nil, fmt.Errorf("compile failed: timed out")
		}
		return nil, fmt.Errorf("compile failed: %s", trimOutput(output+result.Stderr))
	}

	return &dockerProgram{
		run: func(ctx context.Context, input string) (executionResult, error) {
			_, result, err := e.runContainer(ctx, image, workDir, input, e.runtimeMemory, runCmd...)
			return result, err
		},
	}, nil
}

func (e *DockerEvaluator) runContainer(ctx context.Context, image string, workDir string, stdin string, memory string, command ...string) (string, executionResult, error) {
	containerName := fmt.Sprintf("algojudge-%d", time.Now().UnixNano())
	mountArg := fmt.Sprintf("%s:%s", filepath.Clean(workDir), e.workspacePath)
	args := []string{
		"run", "--rm", "--name", containerName,
		"--network", "none",
		"--pids-limit", "128",
		"--read-only",
		"--tmpfs", "/tmp:rw,size=64m",
		"-v", mountArg,
		"-w", e.workspacePath,
	}
	if memory != "" {
		args = append(args, "--memory", memory)
	}
	if e.cpuLimit != "" {
		args = append(args, "--cpus", e.cpuLimit)
	}
	if stdin != "" {
		args = append(args, "-i")
	}
	args = append(args, image)
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

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
		_ = exec.Command("docker", "rm", "-f", containerName).Run()
		return stdout.String(), result, ctx.Err()
	}
	if err == nil {
		return stdout.String(), result, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	return stdout.String(), result, err
}
