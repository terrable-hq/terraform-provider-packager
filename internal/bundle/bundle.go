package bundle

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const defaultOutputDirectory = ".terrable/build"

var validArtifactName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Request describes one Lambda artifact build.
type Request struct {
	Name             string
	Entrypoint       string
	WorkingDirectory string
	OutputDirectory  string
}

// RunRequest describes the JavaScript bundle the runner must produce before it
// is placed in a deterministic Lambda ZIP archive.
type RunRequest struct {
	Entrypoint       string
	OutputFile       string
	WorkingDirectory string
}

// Runner produces the JavaScript bundle consumed by Build.
type Runner interface {
	Run(context.Context, RunRequest) error
}

// Result contains the deployable artifact and the hash expected by
// aws_lambda_function.source_code_hash.
type Result struct {
	ArtifactPath string
	Base64SHA256 string
	Size         int64
}

// Build bundles one entrypoint and writes a deterministic ZIP artifact.
func Build(ctx context.Context, request Request, runner Runner) (Result, error) {
	if runner == nil {
		return Result{}, errors.New("bundle runner is required")
	}
	if !validArtifactName.MatchString(request.Name) {
		return Result{}, fmt.Errorf("artifact name %q must contain only letters, numbers, dots, underscores, or hyphens", request.Name)
	}
	if request.Entrypoint == "" {
		return Result{}, errors.New("entrypoint is required")
	}

	workingDirectory := request.WorkingDirectory
	if workingDirectory == "" {
		workingDirectory = "."
	}
	workingDirectory, err := filepath.Abs(workingDirectory)
	if err != nil {
		return Result{}, fmt.Errorf("resolve working directory: %w", err)
	}

	entrypoint := request.Entrypoint
	if !filepath.IsAbs(entrypoint) {
		entrypoint = filepath.Join(workingDirectory, entrypoint)
	}
	entrypoint, err = filepath.Abs(entrypoint)
	if err != nil {
		return Result{}, fmt.Errorf("resolve entrypoint: %w", err)
	}
	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil {
		return Result{}, fmt.Errorf("inspect entrypoint %q: %w", entrypoint, err)
	}
	if !entrypointInfo.Mode().IsRegular() {
		return Result{}, fmt.Errorf("entrypoint %q is not a regular file", entrypoint)
	}

	outputDirectory := request.OutputDirectory
	if outputDirectory == "" {
		outputDirectory = defaultOutputDirectory
	}
	if !filepath.IsAbs(outputDirectory) {
		outputDirectory = filepath.Join(workingDirectory, outputDirectory)
	}
	outputDirectory, err = filepath.Abs(outputDirectory)
	if err != nil {
		return Result{}, fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", outputDirectory, err)
	}

	stagingDirectory, err := os.MkdirTemp(outputDirectory, ".packager-build-")
	if err != nil {
		return Result{}, fmt.Errorf("create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDirectory)

	bundlePath := filepath.Join(stagingDirectory, "index.js")
	if err := runner.Run(ctx, RunRequest{
		Entrypoint:       entrypoint,
		OutputFile:       bundlePath,
		WorkingDirectory: workingDirectory,
	}); err != nil {
		return Result{}, fmt.Errorf("bundle %q: %w", entrypoint, err)
	}
	if _, err := os.Stat(bundlePath); err != nil {
		return Result{}, fmt.Errorf("inspect generated bundle %q: %w", bundlePath, err)
	}

	artifactPath := filepath.Join(outputDirectory, request.Name+".zip")
	if err := writeDeterministicArchive(bundlePath, artifactPath); err != nil {
		return Result{}, err
	}

	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return Result{}, fmt.Errorf("read artifact %q: %w", artifactPath, err)
	}
	hash := sha256.Sum256(artifact)

	return Result{
		ArtifactPath: artifactPath,
		Base64SHA256: base64.StdEncoding.EncodeToString(hash[:]),
		Size:         int64(len(artifact)),
	}, nil
}

func writeDeterministicArchive(bundlePath, artifactPath string) (returnErr error) {
	temporaryArtifact, err := os.CreateTemp(filepath.Dir(artifactPath), ".packager-artifact-*.zip")
	if err != nil {
		return fmt.Errorf("create temporary artifact: %w", err)
	}
	temporaryPath := temporaryArtifact.Name()
	defer func() {
		_ = temporaryArtifact.Close()
		if returnErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()

	bundleFile, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("open generated bundle: %w", err)
	}
	defer bundleFile.Close()

	archiveWriter := zip.NewWriter(temporaryArtifact)
	header := &zip.FileHeader{Name: "index.js", Method: zip.Deflate}
	header.SetMode(0o644)
	header.Modified = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
	entryWriter, err := archiveWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create archive entry: %w", err)
	}
	if _, err := io.Copy(entryWriter, bundleFile); err != nil {
		return fmt.Errorf("write archive entry: %w", err)
	}
	if err := archiveWriter.Close(); err != nil {
		return fmt.Errorf("finish archive: %w", err)
	}
	if err := temporaryArtifact.Close(); err != nil {
		return fmt.Errorf("close artifact: %w", err)
	}
	if err := os.Rename(temporaryPath, artifactPath); err != nil {
		return fmt.Errorf("replace artifact %q: %w", artifactPath, err)
	}
	return nil
}
