package release_test

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseSnapshot(t *testing.T) {
	if os.Getenv("PACKAGER_RELEASE_SNAPSHOT") != "1" {
		t.Skip("run make release-snapshot to build and verify release artifacts")
	}

	repositoryRoot := filepath.Join("..", "..")
	dist := filepath.Join(repositoryRoot, "dist")
	checksumPaths, err := filepath.Glob(filepath.Join(dist, "terraform-provider-packager_*_SHA256SUMS"))
	if err != nil || len(checksumPaths) != 1 {
		t.Fatalf("expected one checksum file, found %v: %v", checksumPaths, err)
	}
	version := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(checksumPaths[0]), "terraform-provider-packager_"), "_SHA256SUMS")
	prefix := "terraform-provider-packager_" + version
	manifestName := prefix + "_manifest.json"
	expected := map[string]bool{manifestName: true}
	for _, target := range []string{"darwin_amd64", "darwin_arm64", "linux_amd64", "linux_arm64", "windows_amd64", "windows_arm64"} {
		expected[prefix+"_"+target+".zip"] = true
	}

	checksums, err := os.ReadFile(checksumPaths[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(checksums)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !expected[fields[1]] {
			t.Fatalf("unexpected or duplicate checksum entry: %q", line)
		}
		name := fields[1]
		delete(expected, name)
		artifactPath := filepath.Join(dist, name)
		if name == manifestName {
			// GoReleaser uploads this source file with a versioned name. A
			// snapshot skips uploading, so verify the same source bytes here.
			artifactPath = filepath.Join(repositoryRoot, "terraform-registry-manifest.json")
		}
		contents, err := os.ReadFile(artifactPath)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		hash := sha256.Sum256(contents)
		if actual := hex.EncodeToString(hash[:]); actual != fields[0] {
			t.Fatalf("checksum mismatch for %s: got %s, want %s", name, actual, fields[0])
		}
		if name == manifestName {
			checkManifest(t, contents)
			continue
		}
		binaryName := "terraform-provider-packager_v" + version
		if strings.Contains(name, "_windows_") {
			binaryName += ".exe"
		}
		checkArchive(t, artifactPath, binaryName)
	}
	if len(expected) != 0 {
		t.Fatalf("missing checksum entries: %v", expected)
	}
}

func checkManifest(t *testing.T, contents []byte) {
	t.Helper()
	var manifest struct {
		Version  int `json:"version"`
		Metadata struct {
			ProtocolVersions []string `json:"protocol_versions"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatalf("parse Registry manifest: %v", err)
	}
	if manifest.Version != 1 || len(manifest.Metadata.ProtocolVersions) != 1 || manifest.Metadata.ProtocolVersions[0] != "6.0" {
		t.Fatalf("unexpected Registry manifest: %+v", manifest)
	}
}

func checkArchive(t *testing.T, path, binaryName string) {
	t.Helper()
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, entry := range archive.File {
		if entry.Name == binaryName && entry.UncompressedSize64 > 0 {
			return
		}
	}
	t.Fatalf("%s does not contain a non-empty %s at its root", path, binaryName)
}
