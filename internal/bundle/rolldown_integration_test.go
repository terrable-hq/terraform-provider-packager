package bundle

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRolldownBuildsTypeScriptLambdaArtifact(t *testing.T) {
	if os.Getenv("PACKAGER_INTEGRATION") != "1" {
		t.Skip("set PACKAGER_INTEGRATION=1 to run the Rolldown integration test")
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	executableName := "rolldown"
	if runtime.GOOS == "windows" {
		executableName = "rolldown.cmd"
	}
	rolldownPath := filepath.Join(repositoryRoot, "node_modules", ".bin", executableName)

	result, err := Build(context.Background(), Request{
		Name:             "typescript-handler",
		Entrypoint:       "tests/fixtures/basic/src/handler.ts",
		WorkingDirectory: repositoryRoot,
		OutputDirectory:  t.TempDir(),
	}, RolldownRunner{Executable: rolldownPath})
	if err != nil {
		t.Fatalf("Build returned an error: %v", err)
	}

	artifact, err := zip.OpenReader(result.ArtifactPath)
	if err != nil {
		t.Fatalf("open artifact: %v", err)
	}
	defer artifact.Close()
	if len(artifact.File) != 1 {
		t.Fatalf("archive contains %d files, want 1", len(artifact.File))
	}
	entry, err := artifact.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer entry.Close()
	bundleBytes, err := io.ReadAll(entry)
	if err != nil {
		t.Fatal(err)
	}
	bundleSource := string(bundleBytes)
	if !strings.Contains(bundleSource, "handler") || !strings.Contains(bundleSource, "Hello from Rolldown") {
		t.Fatalf("generated bundle did not contain the expected handler:\n%s", bundleSource)
	}
}
