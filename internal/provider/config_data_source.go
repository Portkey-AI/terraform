package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &configDataSource{}
	_ datasource.DataSourceWithConfigure = &configDataSource{}
)

// NewConfigDataSource is a helper function to simplify the provider implementation.
func NewConfigDataSource() datasource.DataSource {
	return &configDataSource{}
}

// configDataSource is the data source implementation.
type configDataSource struct {
	client *client.Client
}

// configDataSourceModel maps the data source schema data.
type configDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Name        types.String `tfsdk:"name"`
	Config      types.String `tfsdk:"config"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Status      types.String `tfsdk:"status"`
	VersionID   types.String `tfsdk:"version_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *configDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the data source.
func (d *configDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Portkey config.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The slug of the config to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Config identifier (UUID).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the config.",
				Computed:    true,
			},
			"config": schema.StringAttribute{
				Description: "JSON configuration object.",
				Computed:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID the config belongs to.",
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this config is the default for the workspace.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the config (active, archived).",
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "Current version ID of the config.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the config was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the config was last updated.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *configDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *configDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state configDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := d.client.GetConfig(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Config",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(config.ID)
	state.Slug = types.StringValue(config.Slug)
	state.Name = types.StringValue(config.Name)
	state.WorkspaceID = types.StringValue(config.WorkspaceID)
	state.IsDefault = types.BoolValue(config.IsDefault == 1)
	state.Status = types.StringValue(config.Status)
	state.VersionID = types.StringValue(config.VersionID)

	// Convert config map to JSON string
	if config.Config != nil {
		configBytes, err := json.Marshal(config.Config)
		if err == nil {
			state.Config = types.StringValue(string(configBytes))
		}
	} else if config.ConfigRaw != "" {
		state.Config = types.StringValue(config.ConfigRaw)
	}

	state.CreatedAt = types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !config.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

