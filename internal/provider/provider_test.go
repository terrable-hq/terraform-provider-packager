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
