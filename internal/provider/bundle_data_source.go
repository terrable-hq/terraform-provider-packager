package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terrable-hq/terraform-provider-packager/internal/bundle"
)

var _ datasource.DataSource = (*bundleDataSource)(nil)

type bundleDataSource struct{}

type bundleDataSourceModel struct {
	Name             types.String `tfsdk:"name"`
	Entrypoint       types.String `tfsdk:"entrypoint"`
	WorkingDirectory types.String `tfsdk:"working_directory"`
	OutputDirectory  types.String `tfsdk:"output_directory"`
	RolldownPath     types.String `tfsdk:"rolldown_path"`
	ArtifactPath     types.String `tfsdk:"artifact_path"`
	Base64SHA256     types.String `tfsdk:"base64sha256"`
	Size             types.Int64  `tfsdk:"size"`
}

// NewBundleDataSource returns the Rolldown bundle data source.
func NewBundleDataSource() datasource.DataSource {
	return &bundleDataSource{}
}

func (d *bundleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundle"
}

func (d *bundleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bundles one JavaScript or TypeScript entrypoint with Rolldown and writes a deterministic Lambda ZIP artifact.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Artifact name. The ZIP file uses this name.",
			},
			"entrypoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "JavaScript or TypeScript entrypoint to bundle.",
			},
			"working_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Directory used to resolve relative entrypoint and output paths. Defaults to the Terraform process working directory.",
			},
			"output_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Directory for generated artifacts. Defaults to `.terrable/build` beneath `working_directory`.",
			},
			"rolldown_path": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Explicit Rolldown executable. By default the provider checks `node_modules/.bin/rolldown` and then PATH.",
			},
			"artifact_path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Absolute path to the generated Lambda ZIP artifact.",
			},
			"base64sha256": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Base64-encoded SHA-256 hash suitable for `aws_lambda_function.source_code_hash`.",
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Artifact size in bytes.",
			},
		},
	}
}

func (d *bundleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bundleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := bundle.Build(ctx, bundle.Request{
		Name:             config.Name.ValueString(),
		Entrypoint:       config.Entrypoint.ValueString(),
		WorkingDirectory: optionalString(config.WorkingDirectory),
		OutputDirectory:  optionalString(config.OutputDirectory),
	}, bundle.RolldownRunner{Executable: optionalString(config.RolldownPath)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to package function", err.Error())
		return
	}

	config.ArtifactPath = types.StringValue(result.ArtifactPath)
	config.Base64SHA256 = types.StringValue(result.Base64SHA256)
	config.Size = types.Int64Value(result.Size)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func optionalString(value types.String) string {
	if value.IsNull() || value.IsUnknown() {
		return ""
	}
	return value.ValueString()
}
