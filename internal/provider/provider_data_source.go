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
	_ datasource.DataSource              = &providerDataSource{}
	_ datasource.DataSourceWithConfigure = &providerDataSource{}
)

// NewProviderDataSource is a helper function to simplify the provider implementation.
func NewProviderDataSource() datasource.DataSource {
	return &providerDataSource{}
}

// providerDataSource is the data source implementation.
type providerDataSource struct {
	client *client.Client
}

// providerDataSourceModel maps the data source schema data.
type providerDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	WorkspaceID   types.String `tfsdk:"workspace_id"`
	IntegrationID types.String `tfsdk:"integration_id"`
	AIProviderID  types.String `tfsdk:"ai_provider_id"`
	Note          types.String `tfsdk:"note"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *providerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the data source.
func (d *providerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Portkey Provider by ID. Requires workspace_id.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Provider identifier (UUID).",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID (UUID) where this provider exists.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the provider.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the provider.",
				Computed:    true,
			},
			"integration_id": schema.StringAttribute{
				Description: "Integration slug or ID used by this provider.",
				Computed:    true,
			},
			"ai_provider_id": schema.StringAttribute{
				Description: "AI provider type (e.g., 'openai', 'anthropic').",
				Computed:    true,
			},
			"note": schema.StringAttribute{
				Description: "Note or description for this provider.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the provider (active, archived).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the provider was created.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *providerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *providerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state providerDataSourceModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get provider from Portkey
	provider, err := d.client.GetProvider(ctx, state.ID.ValueString(), state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Provider",
			err.Error(),
		)
		return
	}

	// Map response to state
	state.Slug = types.StringValue(provider.Slug)
	state.Name = types.StringValue(provider.Name)
	state.Status = types.StringValue(provider.Status)
	state.AIProviderID = types.StringValue(provider.AIProviderID)
	state.CreatedAt = types.StringValue(provider.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if provider.IntegrationID != "" {
		state.IntegrationID = types.StringValue(provider.IntegrationID)
	} else {
		state.IntegrationID = types.StringNull()
	}

	if provider.Note != "" {
		state.Note = types.StringValue(provider.Note)
	} else {
		state.Note = types.StringNull()
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

