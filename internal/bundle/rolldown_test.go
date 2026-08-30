package bundle

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRolldownRunnerUsesExplicitExecutable(t *testing.T) {
	expected := filepath.Join(t.TempDir(), "custom-rolldown")

	actual, err := (RolldownRunner{Executable: expected}).resolveExecutable(t.TempDir())
	if err != nil {
		t.Fatalf("resolveExecutable returned an error: %v", err)
	}
	if actual != expected {
		t.Fatalf("executable = %q, want %q", actual, expected)
	}
}

func TestRolldownRunnerUsesProjectLocalExecutable(t *testing.T) {
	workingDirectory := t.TempDir()
	executableName := "rolldown"
	if runtime.GOOS == "windows" {
		executableName = "rolldown.cmd"
	}
	expected := filepath.Join(workingDirectory, "node_modules", ".bin", executableName)
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(expected, []byte("placeholder"), 0o755); err != nil {
		t.Fatal(err)
	}

	actual, err := (RolldownRunner{}).resolveExecutable(workingDirectory)
	if err != nil {
		t.Fatalf("resolveExecutable returned an error: %v", err)
	}
	if actual != expected {
		t.Fatalf("executable = %q, want %q", actual, expected)
	}
}
