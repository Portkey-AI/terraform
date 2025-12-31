package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/portkey-ai/terraform-provider-portkey/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &usageLimitsPolicyResource{}
	_ resource.ResourceWithConfigure   = &usageLimitsPolicyResource{}
	_ resource.ResourceWithImportState = &usageLimitsPolicyResource{}
)

// NewUsageLimitsPolicyResource is a helper function to simplify the provider implementation.
func NewUsageLimitsPolicyResource() resource.Resource {
	return &usageLimitsPolicyResource{}
}

// usageLimitsPolicyResource is the resource implementation.
type usageLimitsPolicyResource struct {
	client *client.Client
}

// usageLimitsPolicyResourceModel maps the resource schema data.
type usageLimitsPolicyResourceModel struct {
	ID             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	WorkspaceID    types.String  `tfsdk:"workspace_id"`
	Conditions     types.String  `tfsdk:"conditions"`
	GroupBy        types.String  `tfsdk:"group_by"`
	Type           types.String  `tfsdk:"type"`
	CreditLimit    types.Float64 `tfsdk:"credit_limit"`
	AlertThreshold types.Float64 `tfsdk:"alert_threshold"`
	PeriodicReset  types.String  `tfsdk:"periodic_reset"`
	Status         types.String  `tfsdk:"status"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	UpdatedAt      types.String  `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *usageLimitsPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_limits_policy"
}

// Schema defines the schema for the resource.
func (r *usageLimitsPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Portkey usage limits policy. Controls total usage (cost or tokens) over a period.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Policy identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the policy.",
				Optional:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID to create the policy in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"conditions": schema.StringAttribute{
				Description: "JSON array of conditions that define which requests the policy applies to. Each condition has 'key' and 'value'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_by": schema.StringAttribute{
				Description: "JSON array of group by fields that define how usage is aggregated. Each item has 'key'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Policy type: 'cost' or 'tokens'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credit_limit": schema.Float64Attribute{
				Description: "Maximum usage allowed.",
				Required:    true,
			},
			"alert_threshold": schema.Float64Attribute{
				Description: "Threshold at which to send alerts. Must be less than credit_limit.",
				Optional:    true,
			},
			"periodic_reset": schema.StringAttribute{
				Description: "Reset period: 'monthly' or 'weekly'. If not provided, limit is cumulative.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the policy (active, archived).",
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
	}
}

// Configure adds the provider configured client to the resource.
func (r *usageLimitsPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *usageLimitsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan usageLimitsPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse conditions JSON
	var conditions []client.PolicyCondition
	if err := json.Unmarshal([]byte(plan.Conditions.ValueString()), &conditions); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Conditions JSON",
			"The conditions attribute must be a valid JSON array: "+err.Error(),
		)
		return
	}

	// Parse group_by JSON
	var groupBy []client.PolicyGroupBy
	if err := json.Unmarshal([]byte(plan.GroupBy.ValueString()), &groupBy); err != nil {
		resp.Diagnostics.AddError(
			"Invalid GroupBy JSON",
			"The group_by attribute must be a valid JSON array: "+err.Error(),
		)
		return
	}

	// Create new policy
	createReq := client.CreateUsageLimitsPolicyRequest{
		Name:        plan.Name.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Conditions:  conditions,
		GroupBy:     groupBy,
		Type:        plan.Type.ValueString(),
		CreditLimit: plan.CreditLimit.ValueFloat64(),
	}

	if !plan.AlertThreshold.IsNull() && !plan.AlertThreshold.IsUnknown() {
		alertThreshold := plan.AlertThreshold.ValueFloat64()
		createReq.AlertThreshold = &alertThreshold
	}

	if !plan.PeriodicReset.IsNull() && !plan.PeriodicReset.IsUnknown() {
		createReq.PeriodicReset = plan.PeriodicReset.ValueString()
	}

	createResp, err := r.client.CreateUsageLimitsPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating usage limits policy",
			"Could not create policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the full policy details
	policy, err := r.client.GetUsageLimitsPolicy(ctx, createResp.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policy after creation",
			"Could not read policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema
	r.mapPolicyToState(&plan, policy, false)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *usageLimitsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state usageLimitsPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy value from Portkey
	policy, err := r.client.GetUsageLimitsPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Portkey Usage Limits Policy",
			"Could not read policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Preserve user's JSON formatting
	oldConditions := state.Conditions
	oldGroupBy := state.GroupBy

	r.mapPolicyToState(&state, policy, true)

	// Keep original formatting if semantically equal
	state.Conditions = preserveJSONFormatting(oldConditions.ValueString(), state.Conditions.ValueString())
	state.GroupBy = preserveJSONFormatting(oldGroupBy.ValueString(), state.GroupBy.ValueString())

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *usageLimitsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan usageLimitsPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state usageLimitsPolicyResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := client.UpdateUsageLimitsPolicyRequest{}

	if plan.Name.ValueString() != state.Name.ValueString() {
		updateReq.Name = plan.Name.ValueString()
	}

	if plan.CreditLimit.ValueFloat64() != state.CreditLimit.ValueFloat64() {
		creditLimit := plan.CreditLimit.ValueFloat64()
		updateReq.CreditLimit = &creditLimit
	}

	if !plan.AlertThreshold.IsNull() && !plan.AlertThreshold.IsUnknown() {
		if plan.AlertThreshold.ValueFloat64() != state.AlertThreshold.ValueFloat64() {
			alertThreshold := plan.AlertThreshold.ValueFloat64()
			updateReq.AlertThreshold = &alertThreshold
		}
	}

	policy, err := r.client.UpdateUsageLimitsPolicy(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Portkey Usage Limits Policy",
			"Could not update policy, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to plan
	r.mapPolicyToState(&plan, policy, false)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *usageLimitsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state usageLimitsPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing policy
	err := r.client.DeleteUsageLimitsPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey Usage Limits Policy",
			"Could not delete policy, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *usageLimitsPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapPolicyToState maps a UsageLimitsPolicy API response to the Terraform state model
// preserveRequiresReplace controls whether to preserve state values for RequiresReplace attributes
func (r *usageLimitsPolicyResource) mapPolicyToState(state *usageLimitsPolicyResourceModel, policy *client.UsageLimitsPolicy, preserveRequiresReplace bool) {
	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	// Preserve workspace_id from state to avoid triggering RequiresReplace unnecessarily
	if !preserveRequiresReplace || state.WorkspaceID.IsNull() || state.WorkspaceID.IsUnknown() {
		state.WorkspaceID = types.StringValue(policy.WorkspaceID)
	}
	// Preserve type from state to avoid triggering RequiresReplace unnecessarily
	if !preserveRequiresReplace || state.Type.IsNull() || state.Type.IsUnknown() {
		state.Type = types.StringValue(policy.Type)
	}
	state.CreditLimit = types.Float64Value(policy.CreditLimit)
	state.Status = types.StringValue(policy.Status)

	if policy.AlertThreshold != nil {
		state.AlertThreshold = types.Float64Value(*policy.AlertThreshold)
	} else {
		state.AlertThreshold = types.Float64Null()
	}

	// Preserve periodic_reset from state to avoid triggering RequiresReplace unnecessarily
	if !preserveRequiresReplace || state.PeriodicReset.IsNull() || state.PeriodicReset.IsUnknown() {
		if policy.PeriodicReset != "" {
			state.PeriodicReset = types.StringValue(policy.PeriodicReset)
		} else {
			state.PeriodicReset = types.StringNull()
		}
	}

	// Convert conditions to JSON string - preserve from state if set (RequiresReplace)
	if !preserveRequiresReplace || state.Conditions.IsNull() || state.Conditions.IsUnknown() {
		if policy.Conditions != nil {
			conditionsBytes, err := json.Marshal(policy.Conditions)
			if err == nil {
				state.Conditions = types.StringValue(string(conditionsBytes))
			}
		}
	}

	// Convert group_by to JSON string - preserve from state if set (RequiresReplace)
	if !preserveRequiresReplace || state.GroupBy.IsNull() || state.GroupBy.IsUnknown() {
		if policy.GroupBy != nil {
			groupByBytes, err := json.Marshal(policy.GroupBy)
			if err == nil {
				state.GroupBy = types.StringValue(string(groupByBytes))
			}
		}
	}

	state.CreatedAt = types.StringValue(policy.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !policy.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(policy.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
}

// preserveJSONFormatting keeps user's JSON format if semantically equal
func preserveJSONFormatting(oldJSON, newJSON string) types.String {
	if oldJSON == "" {
		return types.StringValue(newJSON)
	}

	var oldVal, newVal interface{}
	oldErr := json.Unmarshal([]byte(oldJSON), &oldVal)
	newErr := json.Unmarshal([]byte(newJSON), &newVal)

	if oldErr == nil && newErr == nil {
		oldBytes, _ := json.Marshal(oldVal)
		newBytes, _ := json.Marshal(newVal)
		if string(oldBytes) == string(newBytes) {
			return types.StringValue(oldJSON)
		}
	}

	return types.StringValue(newJSON)
}
