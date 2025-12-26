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
	_ datasource.DataSource              = &usageLimitsPoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &usageLimitsPoliciesDataSource{}
)

// NewUsageLimitsPoliciesDataSource is a helper function to simplify the provider implementation.
func NewUsageLimitsPoliciesDataSource() datasource.DataSource {
	return &usageLimitsPoliciesDataSource{}
}

// usageLimitsPoliciesDataSource is the data source implementation.
type usageLimitsPoliciesDataSource struct {
	client *client.Client
}

// usageLimitsPoliciesDataSourceModel maps the data source schema data.
type usageLimitsPoliciesDataSourceModel struct {
	WorkspaceID types.String                       `tfsdk:"workspace_id"`
	Policies    []usageLimitsPolicySummaryModel `tfsdk:"policies"`
}

// usageLimitsPolicySummaryModel maps policy summary data.
type usageLimitsPolicySummaryModel struct {
	ID             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	WorkspaceID    types.String  `tfsdk:"workspace_id"`
	Type           types.String  `tfsdk:"type"`
	CreditLimit    types.Float64 `tfsdk:"credit_limit"`
	AlertThreshold types.Float64 `tfsdk:"alert_threshold"`
	PeriodicReset  types.String  `tfsdk:"periodic_reset"`
	Status         types.String  `tfsdk:"status"`
	Conditions     types.String  `tfsdk:"conditions"`
	GroupBy        types.String  `tfsdk:"group_by"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	UpdatedAt      types.String  `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *usageLimitsPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_limits_policies"
}

// Schema defines the schema for the data source.
func (d *usageLimitsPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get a list of Portkey usage limits policies.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID to filter policies.",
				Required:    true,
			},
			"policies": schema.ListNestedAttribute{
				Description: "List of usage limits policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Policy identifier (UUID).",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable name for the policy.",
							Computed:    true,
						},
						"workspace_id": schema.StringAttribute{
							Description: "Workspace ID the policy belongs to.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Policy type: 'cost' or 'tokens'.",
							Computed:    true,
						},
						"credit_limit": schema.Float64Attribute{
							Description: "Maximum usage allowed.",
							Computed:    true,
						},
						"alert_threshold": schema.Float64Attribute{
							Description: "Alert threshold.",
							Computed:    true,
						},
						"periodic_reset": schema.StringAttribute{
							Description: "Reset period.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the policy (active, archived).",
							Computed:    true,
						},
						"conditions": schema.StringAttribute{
							Description: "JSON array of conditions.",
							Computed:    true,
						},
						"group_by": schema.StringAttribute{
							Description: "JSON array of group by fields.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the policy was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the policy was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *usageLimitsPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *usageLimitsPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usageLimitsPoliciesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListUsageLimitsPolicies(ctx, state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Usage Limits Policies",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, policy := range policies {
		policyState := usageLimitsPolicySummaryModel{
			ID:          types.StringValue(policy.ID),
			Name:        types.StringValue(policy.Name),
			WorkspaceID: types.StringValue(policy.WorkspaceID),
			Type:        types.StringValue(policy.Type),
			CreditLimit: types.Float64Value(policy.CreditLimit),
			Status:      types.StringValue(policy.Status),
			CreatedAt:   types.StringValue(policy.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}

		if policy.AlertThreshold != nil {
			policyState.AlertThreshold = types.Float64Value(*policy.AlertThreshold)
		} else {
			policyState.AlertThreshold = types.Float64Null()
		}

		if policy.PeriodicReset != "" {
			policyState.PeriodicReset = types.StringValue(policy.PeriodicReset)
		} else {
			policyState.PeriodicReset = types.StringNull()
		}

		if policy.Conditions != nil {
			conditionsBytes, err := json.Marshal(policy.Conditions)
			if err == nil {
				policyState.Conditions = types.StringValue(string(conditionsBytes))
			}
		}

		if policy.GroupBy != nil {
			groupByBytes, err := json.Marshal(policy.GroupBy)
			if err == nil {
				policyState.GroupBy = types.StringValue(string(groupByBytes))
			}
		}

		if !policy.UpdatedAt.IsZero() {
			policyState.UpdatedAt = types.StringValue(policy.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		state.Policies = append(state.Policies, policyState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

