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
	_ resource.Resource                = &guardrailResource{}
	_ resource.ResourceWithConfigure   = &guardrailResource{}
	_ resource.ResourceWithImportState = &guardrailResource{}
)

// NewGuardrailResource is a helper function to simplify the provider implementation.
func NewGuardrailResource() resource.Resource {
	return &guardrailResource{}
}

// guardrailResource is the resource implementation.
type guardrailResource struct {
	client *client.Client
}

// guardrailResourceModel maps the resource schema data.
type guardrailResourceModel struct {
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

// Metadata returns the resource type name.
func (r *guardrailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_guardrail"
}

// Schema defines the schema for the resource.
func (r *guardrailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Portkey guardrail. Guardrails provide content safety, validation, and policy enforcement for AI requests and responses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Guardrail identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the guardrail. Auto-generated based on name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the guardrail.",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID to create the guardrail in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"checks": schema.StringAttribute{
				Description: "JSON array of guardrail checks. Each check has an 'id' and optional 'parameters'.",
				Required:    true,
			},
			"actions": schema.StringAttribute{
				Description: "JSON object defining actions when checks pass or fail (e.g., onFail, message).",
				Required:    true,
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

// Configure adds the provider configured client to the resource.
func (r *guardrailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *guardrailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan guardrailResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse checks JSON
	var checks []client.GuardrailCheck
	if err := json.Unmarshal([]byte(plan.Checks.ValueString()), &checks); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Checks JSON",
			"The checks attribute must be a valid JSON array: "+err.Error(),
		)
		return
	}

	// Parse actions JSON
	var actions map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Actions.ValueString()), &actions); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Actions JSON",
			"The actions attribute must be a valid JSON object: "+err.Error(),
		)
		return
	}

	// Create new guardrail
	createReq := client.CreateGuardrailRequest{
		Name:        plan.Name.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Checks:      checks,
		Actions:     actions,
	}

	createResp, err := r.client.CreateGuardrail(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating guardrail",
			"Could not create guardrail, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the full guardrail details
	guardrail, err := r.client.GetGuardrail(ctx, createResp.Slug)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading guardrail after creation",
			"Could not read guardrail, unexpected error: "+err.Error(),
		)
		return
	}

	// Keep user's JSON formatting for checks and actions
	oldChecks := plan.Checks
	oldActions := plan.Actions

	// Map response body to schema
	r.mapGuardrailToState(&plan, guardrail, false)

	// Preserve user's formatting if semantically equal (handles is_enabled normalization)
	plan.Checks = r.preserveChecksFormatting(oldChecks.ValueString(), plan.Checks.ValueString())
	plan.Actions = r.preserveJSONFormatting(oldActions.ValueString(), plan.Actions.ValueString())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *guardrailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state guardrailResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed guardrail value from Portkey
	guardrail, err := r.client.GetGuardrail(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Portkey Guardrail",
			"Could not read Portkey guardrail slug "+state.Slug.ValueString()+": "+err.Error(),
		)
		return
	}

	// Keep user's JSON formatting for checks and actions
	oldChecks := state.Checks
	oldActions := state.Actions

	r.mapGuardrailToState(&state, guardrail, true)

	// Compare and preserve formatting if semantically equal (handles is_enabled normalization)
	state.Checks = r.preserveChecksFormatting(oldChecks.ValueString(), state.Checks.ValueString())
	state.Actions = r.preserveJSONFormatting(oldActions.ValueString(), state.Actions.ValueString())

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *guardrailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan guardrailResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for the slug
	var state guardrailResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := client.UpdateGuardrailRequest{}

	// Check what changed
	if plan.Name.ValueString() != state.Name.ValueString() {
		updateReq.Name = plan.Name.ValueString()
	}

	if plan.Checks.ValueString() != state.Checks.ValueString() {
		var checks []client.GuardrailCheck
		if err := json.Unmarshal([]byte(plan.Checks.ValueString()), &checks); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Checks JSON",
				"The checks attribute must be a valid JSON array: "+err.Error(),
			)
			return
		}
		updateReq.Checks = checks
	}

	if plan.Actions.ValueString() != state.Actions.ValueString() {
		var actions map[string]interface{}
		if err := json.Unmarshal([]byte(plan.Actions.ValueString()), &actions); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Actions JSON",
				"The actions attribute must be a valid JSON object: "+err.Error(),
			)
			return
		}
		updateReq.Actions = actions
	}

	guardrail, err := r.client.UpdateGuardrail(ctx, state.Slug.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Portkey Guardrail",
			"Could not update guardrail, unexpected error: "+err.Error(),
		)
		return
	}

	// Keep user's JSON formatting for checks and actions
	oldChecks := plan.Checks
	oldActions := plan.Actions

	// Map response to plan
	r.mapGuardrailToState(&plan, guardrail, false)

	// Preserve user's formatting if semantically equal (handles is_enabled normalization)
	plan.Checks = r.preserveChecksFormatting(oldChecks.ValueString(), plan.Checks.ValueString())
	plan.Actions = r.preserveJSONFormatting(oldActions.ValueString(), plan.Actions.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *guardrailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state guardrailResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing guardrail
	err := r.client.DeleteGuardrail(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey Guardrail",
			"Could not delete guardrail, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *guardrailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by slug
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// mapGuardrailToState maps a Guardrail API response to the Terraform state model
// preserveRequiresReplace controls whether to preserve state values for RequiresReplace attributes
func (r *guardrailResource) mapGuardrailToState(state *guardrailResourceModel, guardrail *client.Guardrail, preserveRequiresReplace bool) {
	state.ID = types.StringValue(guardrail.ID)
	state.Slug = types.StringValue(guardrail.Slug)
	state.Name = types.StringValue(guardrail.Name)
	// Always preserve workspace_id from state if set (API returns UUID but user may have provided slug)
	if state.WorkspaceID.IsNull() || state.WorkspaceID.IsUnknown() {
		state.WorkspaceID = types.StringValue(guardrail.WorkspaceID)
	}
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
}

// preserveJSONFormatting keeps user's JSON format if semantically equal
func (r *guardrailResource) preserveJSONFormatting(oldJSON, newJSON string) types.String {
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

// preserveChecksFormatting handles is_enabled normalization for guardrail checks
// The API may omit is_enabled when it's true (the default), but the user may have explicitly set it
func (r *guardrailResource) preserveChecksFormatting(oldJSON, newJSON string) types.String {
	if oldJSON == "" {
		return types.StringValue(newJSON)
	}

	var oldChecks, newChecks []map[string]interface{}
	oldErr := json.Unmarshal([]byte(oldJSON), &oldChecks)
	newErr := json.Unmarshal([]byte(newJSON), &newChecks)

	if oldErr != nil || newErr != nil {
		return types.StringValue(newJSON)
	}

	// Normalize is_enabled: remove if true (which is the default)
	normalizeChecks := func(checks []map[string]interface{}) {
		for _, check := range checks {
			if isEnabled, ok := check["is_enabled"]; ok {
				// Remove is_enabled if it's true (the default value)
				if enabled, isBool := isEnabled.(bool); isBool && enabled {
					delete(check, "is_enabled")
				}
			}
		}
	}

	normalizeChecks(oldChecks)
	normalizeChecks(newChecks)

	oldBytes, _ := json.Marshal(oldChecks)
	newBytes, _ := json.Marshal(newChecks)

	if string(oldBytes) == string(newBytes) {
		return types.StringValue(oldJSON)
	}

	return types.StringValue(newJSON)
}
