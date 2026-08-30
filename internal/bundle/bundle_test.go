package bundle

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

type fakeRunner struct {
	request RunRequest
	content string
}

func (r *fakeRunner) Run(_ context.Context, request RunRequest) error {
	r.request = request
	return os.WriteFile(request.OutputFile, []byte(r.content), 0o644)
}

func TestBuildUsesTerrableBuildDirectoryByDefault(t *testing.T) {
	workingDirectory := t.TempDir()
	entrypoint := filepath.Join(workingDirectory, "src", "handler.ts")
	if err := os.MkdirAll(filepath.Dir(entrypoint), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entrypoint, []byte("export const handler = () => 'ok'"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &fakeRunner{content: "exports.handler = () => 'ok';\n"}
	result, err := Build(context.Background(), Request{
		Name:             "hello",
		Entrypoint:       "src/handler.ts",
		WorkingDirectory: workingDirectory,
	}, runner)
	if err != nil {
		t.Fatalf("Build returned an error: %v", err)
	}

	expectedArtifact := filepath.Join(workingDirectory, ".terrable", "build", "hello.zip")
	if result.ArtifactPath != expectedArtifact {
		t.Fatalf("artifact path = %q, want %q", result.ArtifactPath, expectedArtifact)
	}
	if runner.request.Entrypoint != entrypoint {
		t.Fatalf("runner entrypoint = %q, want %q", runner.request.Entrypoint, entrypoint)
	}
	if result.Base64SHA256 == "" {
		t.Fatal("expected a base64 SHA-256 hash")
	}

	archive, err := zip.OpenReader(result.ArtifactPath)
	if err != nil {
		t.Fatalf("open artifact: %v", err)
	}
	defer archive.Close()
	if len(archive.File) != 1 || archive.File[0].Name != "index.js" {
		t.Fatalf("archive entries = %#v, want one index.js entry", archive.File)
	}
}

func TestBuildUsesConfiguredOutputDirectory(t *testing.T) {
	workingDirectory := t.TempDir()
	entrypoint := filepath.Join(workingDirectory, "handler.ts")
	if err := os.WriteFile(entrypoint, []byte("export const handler = () => 'ok'"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Build(context.Background(), Request{
		Name:             "custom",
		Entrypoint:       "handler.ts",
		WorkingDirectory: workingDirectory,
		OutputDirectory:  "artifacts/lambda",
	}, &fakeRunner{content: "exports.handler = () => 'ok';\n"})
	if err != nil {
		t.Fatalf("Build returned an error: %v", err)
	}

	expectedArtifact := filepath.Join(workingDirectory, "artifacts", "lambda", "custom.zip")
	if result.ArtifactPath != expectedArtifact {
		t.Fatalf("artifact path = %q, want %q", result.ArtifactPath, expectedArtifact)
	}
}

func TestBuildProducesDeterministicArchive(t *testing.T) {
	workingDirectory := t.TempDir()
	entrypoint := filepath.Join(workingDirectory, "handler.ts")
	if err := os.WriteFile(entrypoint, []byte("export const handler = () => 'ok'"), 0o644); err != nil {
		t.Fatal(err)
	}

	request := Request{
		Name:             "deterministic",
		Entrypoint:       entrypoint,
		WorkingDirectory: workingDirectory,
	}
	first, err := Build(context.Background(), request, &fakeRunner{content: "exports.handler = () => 'ok';\n"})
	if err != nil {
		t.Fatalf("first Build returned an error: %v", err)
	}
	firstBytes, err := os.ReadFile(first.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Build(context.Background(), request, &fakeRunner{content: "exports.handler = () => 'ok';\n"})
	if err != nil {
		t.Fatalf("second Build returned an error: %v", err)
	}
	secondBytes, err := os.ReadFile(second.ArtifactPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(firstBytes) != string(secondBytes) {
		t.Fatal("repeated builds produced different archive bytes")
	}
	if first.Base64SHA256 != second.Base64SHA256 {
		t.Fatalf("archive hashes differ: %q != %q", first.Base64SHA256, second.Base64SHA256)
	}
}
