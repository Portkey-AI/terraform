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
	_ datasource.DataSource              = &guardrailsDataSource{}
	_ datasource.DataSourceWithConfigure = &guardrailsDataSource{}
)

// NewGuardrailsDataSource is a helper function to simplify the provider implementation.
func NewGuardrailsDataSource() datasource.DataSource {
	return &guardrailsDataSource{}
}

// guardrailsDataSource is the data source implementation.
type guardrailsDataSource struct {
	client *client.Client
}

// guardrailsDataSourceModel maps the data source schema data.
type guardrailsDataSourceModel struct {
	WorkspaceID types.String             `tfsdk:"workspace_id"`
	Guardrails  []guardrailSummaryModel `tfsdk:"guardrails"`
}

// guardrailSummaryModel maps guardrail summary data.
type guardrailSummaryModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Name        types.String `tfsdk:"name"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *guardrailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guardrails"
}

// Schema defines the schema for the data source.
func (d *guardrailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get a list of Portkey guardrails.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID to filter guardrails. Required due to API permission requirements.",
				Required:    true,
			},
			"guardrails": schema.ListNestedAttribute{
				Description: "List of guardrails.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Guardrail identifier (UUID).",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "URL-friendly identifier for the guardrail.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable name for the guardrail.",
							Computed:    true,
						},
						"workspace_id": schema.StringAttribute{
							Description: "Workspace ID the guardrail belongs to.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the guardrail (active, archived).",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the guardrail was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the guardrail was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *guardrailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *guardrailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state guardrailsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	guardrails, err := d.client.ListGuardrails(ctx, state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Guardrails",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, guardrail := range guardrails {
		guardrailState := guardrailSummaryModel{
			ID:          types.StringValue(guardrail.ID),
			Slug:        types.StringValue(guardrail.Slug),
			Name:        types.StringValue(guardrail.Name),
			WorkspaceID: types.StringValue(guardrail.WorkspaceID),
			Status:      types.StringValue(guardrail.Status),
			CreatedAt:   types.StringValue(guardrail.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}

		if !guardrail.UpdatedAt.IsZero() {
			guardrailState.UpdatedAt = types.StringValue(guardrail.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		state.Guardrails = append(state.Guardrails, guardrailState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

