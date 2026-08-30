package bundle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RolldownRunner invokes a project-installed or explicitly configured
// Rolldown executable.
type RolldownRunner struct {
	Executable string
}

// Run bundles the entrypoint for the AWS Lambda Node.js runtime.
func (r RolldownRunner) Run(ctx context.Context, request RunRequest) error {
	executable, err := r.resolveExecutable(request.WorkingDirectory)
	if err != nil {
		return err
	}

	command := exec.CommandContext(
		ctx,
		executable,
		request.Entrypoint,
		"--file", request.OutputFile,
		"--format", "cjs",
		"--platform", "node",
	)
	command.Dir = request.WorkingDirectory
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("run Rolldown: %s", message)
	}
	return nil
}

func (r RolldownRunner) resolveExecutable(workingDirectory string) (string, error) {
	if r.Executable != "" {
		return r.Executable, nil
	}

	executableName := "rolldown"
	if runtime.GOOS == "windows" {
		executableName = "rolldown.cmd"
	}
	projectExecutable := filepath.Join(workingDirectory, "node_modules", ".bin", executableName)
	if info, err := os.Stat(projectExecutable); err == nil && !info.IsDir() {
		return projectExecutable, nil
	}

	pathExecutable, err := exec.LookPath("rolldown")
	if err == nil {
		return pathExecutable, nil
	}
	if !errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("find Rolldown executable: %w", err)
	}
	return "", fmt.Errorf("Rolldown was not found; install it in %q or set rolldown_path", filepath.Join(workingDirectory, "node_modules"))
}
