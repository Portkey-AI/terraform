package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &configsDataSource{}
	_ datasource.DataSourceWithConfigure = &configsDataSource{}
)

// NewConfigsDataSource is a helper function to simplify the provider implementation.
func NewConfigsDataSource() datasource.DataSource {
	return &configsDataSource{}
}

// configsDataSource is the data source implementation.
type configsDataSource struct {
	client *client.Client
}

// configsDataSourceModel maps the data source schema data.
type configsDataSourceModel struct {
	WorkspaceID types.String         `tfsdk:"workspace_id"`
	Configs     []configSummaryModel `tfsdk:"configs"`
}

// configSummaryModel maps config summary data.
type configSummaryModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Name        types.String `tfsdk:"name"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *configsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configs"
}

// Schema defines the schema for the data source.
func (d *configsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get a list of Portkey configs.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "Optional workspace ID to filter configs.",
				Optional:    true,
			},
			"configs": schema.ListNestedAttribute{
				Description: "List of configs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Config identifier (UUID).",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "URL-friendly identifier for the config.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable name for the config.",
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
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the config was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the config was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *configsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *configsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state configsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := ""
	if !state.WorkspaceID.IsNull() {
		workspaceID = state.WorkspaceID.ValueString()
	}

	configs, err := d.client.ListConfigs(ctx, workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Configs",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, config := range configs {
		configState := configSummaryModel{
			ID:          types.StringValue(config.ID),
			Slug:        types.StringValue(config.Slug),
			Name:        types.StringValue(config.Name),
			WorkspaceID: types.StringValue(config.WorkspaceID),
			IsDefault:   types.BoolValue(config.IsDefault == 1),
			Status:      types.StringValue(config.Status),
			CreatedAt:   types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}

		if !config.UpdatedAt.IsZero() {
			configState.UpdatedAt = types.StringValue(config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		state.Configs = append(state.Configs, configState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
