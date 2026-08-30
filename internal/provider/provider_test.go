package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderExposesBundleDataSource(t *testing.T) {
	packagerProvider := New("test")()

	metadata := &frameworkprovider.MetadataResponse{}
	packagerProvider.Metadata(context.Background(), frameworkprovider.MetadataRequest{}, metadata)
	if metadata.TypeName != "packager" {
		t.Fatalf("provider type name = %q, want %q", metadata.TypeName, "packager")
	}

	dataSources := packagerProvider.DataSources(context.Background())
	if len(dataSources) != 1 {
		t.Fatalf("data source count = %d, want 1", len(dataSources))
	}

	dataSourceMetadata := &datasource.MetadataResponse{}
	dataSources[0]().Metadata(
		context.Background(),
		datasource.MetadataRequest{ProviderTypeName: "packager"},
		dataSourceMetadata,
	)
	if dataSourceMetadata.TypeName != "packager_bundle" {
		t.Fatalf("data source type name = %q, want %q", dataSourceMetadata.TypeName, "packager_bundle")
	}
}

func TestBundleDataSourceSchemaContract(t *testing.T) {
	response := &datasource.SchemaResponse{}
	NewBundleDataSource().Schema(context.Background(), datasource.SchemaRequest{}, response)
	if response.Diagnostics.HasError() {
		t.Fatalf("schema returned diagnostics: %v", response.Diagnostics)
	}

	required := []string{"name", "entrypoint"}
	for _, name := range required {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			t.Fatalf("schema is missing required attribute %q", name)
		}
		if !attribute.IsRequired() {
			t.Fatalf("attribute %q is not required", name)
		}
	}

	optional := []string{"working_directory", "output_directory", "rolldown_path"}
	for _, name := range optional {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			t.Fatalf("schema is missing optional attribute %q", name)
		}
		if !attribute.IsOptional() {
			t.Fatalf("attribute %q is not optional", name)
		}
	}

	computed := []string{"artifact_path", "base64sha256", "size"}
	for _, name := range computed {
		attribute, ok := response.Schema.Attributes[name]
		if !ok {
			t.Fatalf("schema is missing computed attribute %q", name)
		}
		if !attribute.IsComputed() {
			t.Fatalf("attribute %q is not computed", name)
		}
	}
}
