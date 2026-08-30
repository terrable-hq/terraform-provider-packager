package bundle

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEmbeddedBundleExecutesInNode(t *testing.T) {
	if os.Getenv("PACKAGER_INTEGRATION") != "1" {
		t.Skip("set PACKAGER_INTEGRATION=1 to execute the generated Lambda handler in Node.js")
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// Node is only used to execute the result, never to build it.
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("Node.js is required only for this runtime smoke test")
	}
	t.Setenv("PATH", t.TempDir())

	result, err := Build(context.Background(), Request{
		Name:             "typescript-handler",
		Entrypoint:       "tests/fixtures/basic/src/handler.ts",
		WorkingDirectory: repositoryRoot,
		OutputDirectory:  t.TempDir(),
	}, EsbuildRunner{})
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
	bundlePath := filepath.Join(t.TempDir(), "index.cjs")
	if err := os.WriteFile(bundlePath, bundleBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(node, "-e", `
const assert = require("node:assert/strict");
require(process.argv[1]).handler().then(result => {
  assert.deepEqual(result, {statusCode: 200, body: "Hello from Terrable"});
  console.log("PASS");
}).catch(error => { console.error(error); process.exitCode = 1; });
`, bundlePath).CombinedOutput()
	if err != nil || string(output) != "PASS\n" {
		t.Fatalf("generated handler failed: %s, %v", output, err)
	}
}
