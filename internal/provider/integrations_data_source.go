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
	_ datasource.DataSource              = &integrationsDataSource{}
	_ datasource.DataSourceWithConfigure = &integrationsDataSource{}
)

// NewIntegrationsDataSource is a helper function to simplify the provider implementation.
func NewIntegrationsDataSource() datasource.DataSource {
	return &integrationsDataSource{}
}

// integrationsDataSource is the data source implementation.
type integrationsDataSource struct {
	client *client.Client
}

// integrationsDataSourceModel maps the data source schema data.
type integrationsDataSourceModel struct {
	Integrations []integrationModel `tfsdk:"integrations"`
}

// integrationModel maps integration data
type integrationModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	AIProviderID types.String `tfsdk:"ai_provider_id"`
	Description  types.String `tfsdk:"description"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *integrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

// Schema defines the schema for the data source.
func (d *integrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Portkey integrations in the organization.",
		Attributes: map[string]schema.Attribute{
			"integrations": schema.ListNestedAttribute{
				Description: "List of integrations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Integration UUID.",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "Integration slug identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable name of the integration.",
							Computed:    true,
						},
						"ai_provider_id": schema.StringAttribute{
							Description: "ID of the AI provider.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the integration.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the integration.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the integration was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the integration was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *integrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *integrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state integrationsDataSourceModel

	// Get integrations from Portkey API
	integrations, err := d.client.ListIntegrations(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Integrations",
			err.Error(),
		)
		return
	}

	// Map response to state
	for _, integration := range integrations {
		integrationState := integrationModel{
			ID:           types.StringValue(integration.ID),
			Slug:         types.StringValue(integration.Slug),
			Name:         types.StringValue(integration.Name),
			AIProviderID: types.StringValue(integration.AIProviderID),
			Status:       types.StringValue(integration.Status),
			CreatedAt:    types.StringValue(integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}
		if integration.Description != "" {
			integrationState.Description = types.StringValue(integration.Description)
		} else {
			integrationState.Description = types.StringNull()
		}
		if !integration.UpdatedAt.IsZero() {
			integrationState.UpdatedAt = types.StringValue(integration.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		} else {
			integrationState.UpdatedAt = types.StringNull()
		}
		state.Integrations = append(state.Integrations, integrationState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

