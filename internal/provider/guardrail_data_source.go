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
	_ datasource.DataSource              = &guardrailDataSource{}
	_ datasource.DataSourceWithConfigure = &guardrailDataSource{}
)

// NewGuardrailDataSource is a helper function to simplify the provider implementation.
func NewGuardrailDataSource() datasource.DataSource {
	return &guardrailDataSource{}
}

// guardrailDataSource is the data source implementation.
type guardrailDataSource struct {
	client *client.Client
}

// guardrailDataSourceModel maps the data source schema data.
type guardrailDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Slug        types.String `tfsdk:"slug"`
	Name        types.String `tfsdk:"name"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Checks      types.String `tfsdk:"checks"`
	Actions     types.String `tfsdk:"actions"`
	Status      types.String `tfsdk:"status"`
	VersionID   types.String `tfsdk:"version_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *guardrailDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guardrail"
}

// Schema defines the schema for the data source.
func (d *guardrailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Portkey guardrail.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The slug of the guardrail to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Guardrail identifier (UUID).",
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
			"checks": schema.StringAttribute{
				Description: "JSON array of guardrail checks.",
				Computed:    true,
			},
			"actions": schema.StringAttribute{
				Description: "JSON object defining actions.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the guardrail (active, archived).",
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "Current version ID of the guardrail.",
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *guardrailDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *guardrailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state guardrailDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	guardrail, err := d.client.GetGuardrail(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Guardrail",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(guardrail.ID)
	state.Slug = types.StringValue(guardrail.Slug)
	state.Name = types.StringValue(guardrail.Name)
	state.WorkspaceID = types.StringValue(guardrail.WorkspaceID)
	state.Status = types.StringValue(guardrail.Status)
	state.VersionID = types.StringValue(guardrail.VersionID)

	// Convert checks to JSON string
	if guardrail.Checks != nil {
		checksBytes, err := json.Marshal(guardrail.Checks)
		if err == nil {
			state.Checks = types.StringValue(string(checksBytes))
		}
	}

	// Convert actions to JSON string
	if guardrail.Actions != nil {
		actionsBytes, err := json.Marshal(guardrail.Actions)
		if err == nil {
			state.Actions = types.StringValue(string(actionsBytes))
		}
	}

	state.CreatedAt = types.StringValue(guardrail.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !guardrail.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(guardrail.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
