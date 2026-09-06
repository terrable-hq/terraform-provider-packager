package bundle

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSource(t *testing.T, root, name, source string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readBundle(t *testing.T, path string) string {
	t.Helper()
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	if len(archive.File) != 1 || archive.File[0].Name != "index.js" {
		t.Fatalf("expected only index.js, got %v", archive.File)
	}
	file, err := archive.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func TestEmbeddedBuildWithoutExternalTools(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	root := t.TempDir()
	writeSource(t, root, "src/handler.ts", `
import { basename } from "node:path";
import { message } from "./message";
import { suffix } from "fixture-dependency";
export const handler = (): string => basename(message + suffix);
`)
	writeSource(t, root, "src/message.ts", `export const message: string = "Hello embedded";`)
	writeSource(t, root, "node_modules/fixture-dependency/package.json", `{"name":"fixture-dependency","main":"index.js"}`)
	writeSource(t, root, "node_modules/fixture-dependency/index.js", `exports.suffix = " dependency bundled";`)
	request := Request{Name: "handler", Entrypoint: "src/handler.ts", WorkingDirectory: root}
	first, err := Build(context.Background(), request, EsbuildRunner{})
	if err != nil {
		t.Fatal(err)
	}
	source := readBundle(t, first.ArtifactPath)
	for _, expected := range []string{"module.exports", "Hello embedded", "dependency bundled", `require("node:path")`} {
		if !strings.Contains(source, expected) {
			t.Errorf("bundle lacks %q:\n%s", expected, source)
		}
	}
	for _, unexpected := range []string{": string", `require("fixture-dependency")`, `require("./message")`} {
		if strings.Contains(source, unexpected) {
			t.Errorf("bundle still contains %q", unexpected)
		}
	}
	second, err := Build(context.Background(), request, EsbuildRunner{})
	if err != nil {
		t.Fatal(err)
	}
	if first.Base64SHA256 != second.Base64SHA256 {
		t.Fatal("real bundler produced a different ZIP hash on a repeated build")
	}
	entries, err := os.ReadDir(filepath.Join(root, ".terrable", "build"))
	if err != nil || len(entries) != 1 || entries[0].Name() != "handler.zip" {
		t.Fatalf("unexpected generated artifacts: %v, %v", entries, err)
	}
	entries, err = os.ReadDir(filepath.Join(root, "src"))
	if err != nil || len(entries) != 2 {
		t.Fatalf("generated files leaked into source directory: %v, %v", entries, err)
	}
}

func TestEmbeddedBuildReportsErrorsAndPreservesArtifact(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, test := range []struct{ name, source, diagnostic string }{
		{"syntax", `export const handler = ;`, "Unexpected"},
		{"missing-dependency", `import value from "missing-package"; export const handler = () => value;`, "Could not resolve"},
		{"extra-output", `import "./style.css"; export const handler = () => "ok";`, "single JavaScript"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeSource(t, root, "src/handler.ts", test.source)
			writeSource(t, root, "src/style.css", "body { color: red }")
			writeSource(t, root, ".terrable/build/handler.zip", "previous artifact")
			_, err := Build(context.Background(), Request{
				Name: "handler", Entrypoint: "src/handler.ts", WorkingDirectory: root,
			}, EsbuildRunner{})
			if err == nil || !strings.Contains(err.Error(), test.diagnostic) {
				t.Fatalf("error = %v, want %q", err, test.diagnostic)
			}
			if test.name != "extra-output" && !strings.Contains(err.Error(), "src/handler.ts:1:") {
				t.Fatalf("diagnostic missing source location: %v", err)
			}
			contents, err := os.ReadFile(filepath.Join(root, ".terrable/build/handler.zip"))
			if err != nil || string(contents) != "previous artifact" {
				t.Fatalf("failed build replaced previous artifact: %q, %v", contents, err)
			}
			entries, err := os.ReadDir(filepath.Join(root, ".terrable/build"))
			if err != nil || len(entries) != 1 {
				t.Fatalf("failed build left staging files: %v, %v", entries, err)
			}
		})
	}
}

func TestEmbeddedBuildHonoursCancellation(t *testing.T) {
	root := t.TempDir()
	writeSource(t, root, "handler.ts", `export const handler = () => "ok";`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Build(ctx, Request{Name: "handler", Entrypoint: "handler.ts", WorkingDirectory: root}, EsbuildRunner{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context cancellation", err)
	}
}
