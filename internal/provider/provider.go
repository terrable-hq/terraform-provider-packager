package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*packagerProvider)(nil)

type packagerProvider struct {
	version string
}

// New returns a Terraform provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &packagerProvider{version: version}
	}
}

func (p *packagerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "packager"
	resp.Version = p.version
}

func (p *packagerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *packagerProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *packagerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewBundleDataSource}
}

func (p *packagerProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
