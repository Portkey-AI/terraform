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
	_ datasource.DataSource              = &providersDataSource{}
	_ datasource.DataSourceWithConfigure = &providersDataSource{}
)

// NewProvidersDataSource is a helper function to simplify the provider implementation.
func NewProvidersDataSource() datasource.DataSource {
	return &providersDataSource{}
}

// providersDataSource is the data source implementation.
type providersDataSource struct {
	client *client.Client
}

// providersDataSourceModel maps the data source schema data.
type providersDataSourceModel struct {
	WorkspaceID types.String            `tfsdk:"workspace_id"`
	Providers   []providerDataItemModel `tfsdk:"providers"`
}

// providerDataItemModel maps individual provider data.
type providerDataItemModel struct {
	ID            types.String `tfsdk:"id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	IntegrationID types.String `tfsdk:"integration_id"`
	AIProviderID  types.String `tfsdk:"ai_provider_id"`
	Note          types.String `tfsdk:"note"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *providersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_providers"
}

// Schema defines the schema for the data source.
func (d *providersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Portkey Providers for a workspace.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID (UUID) to list providers from. Required.",
				Required:    true,
			},
			"providers": schema.ListNestedAttribute{
				Description: "List of providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Provider identifier (UUID).",
							Computed:    true,
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *providersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *providersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state providersDataSourceModel

	// Get config
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get providers from Portkey
	providers, err := d.client.ListProviders(ctx, state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Providers",
			err.Error(),
		)
		return
	}

	// Map response to state
	state.Providers = make([]providerDataItemModel, 0, len(providers))
	for _, p := range providers {
		item := providerDataItemModel{
			ID:           types.StringValue(p.ID),
			Slug:         types.StringValue(p.Slug),
			Name:         types.StringValue(p.Name),
			Status:       types.StringValue(p.Status),
			AIProviderID: types.StringValue(p.AIProviderID),
			CreatedAt:    types.StringValue(p.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}

		if p.IntegrationID != "" {
			item.IntegrationID = types.StringValue(p.IntegrationID)
		} else {
			item.IntegrationID = types.StringNull()
		}

		if p.Note != "" {
			item.Note = types.StringValue(p.Note)
		} else {
			item.Note = types.StringNull()
		}

		state.Providers = append(state.Providers, item)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
