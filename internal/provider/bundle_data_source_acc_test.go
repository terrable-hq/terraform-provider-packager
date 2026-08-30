package provider

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccBundleDataSourceDefaultOutput(t *testing.T) {
	workingDirectory, rolldownPath := prepareAcceptanceFixture(t)

	expectedArtifact := filepath.Join(workingDirectory, ".terrable", "build", "acceptance.zip")
	config := acceptanceBundleConfig(workingDirectory, rolldownPath, "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptanceProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  acceptanceBundleChecks(expectedArtifact),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

func TestAccBundleDataSourceCustomOutput(t *testing.T) {
	workingDirectory, rolldownPath := prepareAcceptanceFixture(t)

	expectedArtifact := filepath.Join(workingDirectory, "artifacts", "lambda", "acceptance.zip")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptanceProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acceptanceBundleConfig(workingDirectory, rolldownPath, "artifacts/lambda"),
				Check:  acceptanceBundleChecks(expectedArtifact),
			},
		},
	})
}

func prepareAcceptanceFixture(t *testing.T) (string, string) {
	t.Helper()

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	workingDirectory := t.TempDir()
	copyAcceptanceFixture(t, repositoryRoot, workingDirectory)

	executableName := "rolldown"
	if runtime.GOOS == "windows" {
		executableName = "rolldown.cmd"
	}
	rolldownPath := filepath.Join(repositoryRoot, "node_modules", ".bin", executableName)
	if _, err := os.Stat(rolldownPath); err != nil {
		t.Fatalf("Rolldown is required for acceptance tests: %v", err)
	}

	return workingDirectory, rolldownPath
}

func acceptanceProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"packager": providerserver.NewProtocol6WithError(New("acceptance")()),
	}
}

func acceptanceBundleConfig(workingDirectory, rolldownPath, outputDirectory string) string {
	outputConfiguration := ""
	if outputDirectory != "" {
		outputConfiguration = fmt.Sprintf("  output_directory  = %q\n", outputDirectory)
	}

	return fmt.Sprintf(`
data "packager_bundle" "handler" {
  name              = "acceptance"
  entrypoint        = "src/handler.ts"
  working_directory = %q
  rolldown_path     = %q
%s
}

resource "terraform_data" "bundle" {
  input = {
    artifact_path = data.packager_bundle.handler.artifact_path
    base64sha256   = data.packager_bundle.handler.base64sha256
    size           = data.packager_bundle.handler.size
  }
}
`, filepath.ToSlash(workingDirectory), filepath.ToSlash(rolldownPath), outputConfiguration)
}

func acceptanceBundleChecks(expectedArtifact string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.packager_bundle.handler",
			"artifact_path",
			expectedArtifact,
		),
		resource.TestCheckResourceAttrWith(
			"data.packager_bundle.handler",
			"artifact_path",
			checkAcceptanceArtifact,
		),
		resource.TestCheckResourceAttrWith(
			"data.packager_bundle.handler",
			"base64sha256",
			checkAcceptanceHash(expectedArtifact),
		),
		resource.TestCheckResourceAttrWith(
			"data.packager_bundle.handler",
			"size",
			checkAcceptanceSize(expectedArtifact),
		),
	)
}

func copyAcceptanceFixture(t *testing.T, repositoryRoot, workingDirectory string) {
	t.Helper()

	for _, filename := range []string{"handler.ts", "message.ts"} {
		sourcePath := filepath.Join(repositoryRoot, "tests", "fixtures", "basic", "src", filename)
		contents, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read fixture %q: %v", sourcePath, err)
		}
		destinationPath := filepath.Join(workingDirectory, "src", filename)
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destinationPath, contents, 0o644); err != nil {
			t.Fatalf("write fixture %q: %v", destinationPath, err)
		}
	}
}

func checkAcceptanceArtifact(path string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open generated artifact: %w", err)
	}
	defer archive.Close()

	if len(archive.File) != 1 || archive.File[0].Name != "index.js" {
		return fmt.Errorf("archive entries = %#v, want one index.js entry", archive.File)
	}
	entry, err := archive.File[0].Open()
	if err != nil {
		return fmt.Errorf("open index.js: %w", err)
	}
	defer entry.Close()
	contents, err := io.ReadAll(entry)
	if err != nil {
		return fmt.Errorf("read index.js: %w", err)
	}
	if !strings.Contains(string(contents), "Hello from Rolldown") {
		return fmt.Errorf("index.js does not contain the bundled fixture")
	}
	return nil
}

func checkAcceptanceHash(artifactPath string) resource.CheckResourceAttrWithFunc {
	return func(actual string) error {
		contents, err := os.ReadFile(artifactPath)
		if err != nil {
			return fmt.Errorf("read generated artifact: %w", err)
		}
		hash := sha256.Sum256(contents)
		expected := base64.StdEncoding.EncodeToString(hash[:])
		if actual != expected {
			return fmt.Errorf("base64sha256 = %q, want %q", actual, expected)
		}
		return nil
	}
}

func checkAcceptanceSize(artifactPath string) resource.CheckResourceAttrWithFunc {
	return func(actual string) error {
		actualSize, err := strconv.ParseInt(actual, 10, 64)
		if err != nil {
			return fmt.Errorf("parse size %q: %w", actual, err)
		}
		info, err := os.Stat(artifactPath)
		if err != nil {
			return fmt.Errorf("inspect generated artifact: %w", err)
		}
		if actualSize != info.Size() {
			return fmt.Errorf("size = %d, want %d", actualSize, info.Size())
		}
		return nil
	}
}
