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
	_ resource.Resource                = &configResource{}
	_ resource.ResourceWithConfigure   = &configResource{}
	_ resource.ResourceWithImportState = &configResource{}
)

// NewConfigResource is a helper function to simplify the provider implementation.
func NewConfigResource() resource.Resource {
	return &configResource{}
}

// configResource is the resource implementation.
type configResource struct {
	client *client.Client
}

// configResourceModel maps the resource schema data.
type configResourceModel struct {
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

// Metadata returns the resource type name.
func (r *configResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the resource.
func (r *configResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Portkey config. Configs define routing rules, caching, retry policies, and other settings for AI requests.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Config identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the config. Auto-generated based on name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the config.",
				Required:    true,
			},
			"config": schema.StringAttribute{
				Description: "JSON configuration object containing routing rules, cache settings, retry policies, etc.",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID to create the config in. Required when using org-level API keys.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this config is the default for the workspace.",
				Optional:    true,
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

// Configure adds the provider configured client to the resource.
func (r *configResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *configResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the config JSON string to a map
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Config.ValueString()), &configMap); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Config JSON",
			"The config attribute must be valid JSON: "+err.Error(),
		)
		return
	}

	// Create new config
	createReq := client.CreateConfigRequest{
		Name:   plan.Name.ValueString(),
		Config: configMap,
	}

	if !plan.WorkspaceID.IsNull() && !plan.WorkspaceID.IsUnknown() {
		createReq.WorkspaceID = plan.WorkspaceID.ValueString()
	}

	if !plan.IsDefault.IsNull() && !plan.IsDefault.IsUnknown() {
		isDefault := 0
		if plan.IsDefault.ValueBool() {
			isDefault = 1
		}
		createReq.IsDefault = &isDefault
	}

	createResp, err := r.client.CreateConfig(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating config",
			"Could not create config, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the full config details
	config, err := r.client.GetConfig(ctx, createResp.Slug)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading config after creation",
			"Could not read config, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema
	plan.ID = types.StringValue(config.ID)
	plan.Slug = types.StringValue(config.Slug)
	plan.WorkspaceID = types.StringValue(config.WorkspaceID)
	plan.IsDefault = types.BoolValue(config.IsDefault == 1)
	plan.Status = types.StringValue(config.Status)
	plan.VersionID = types.StringValue(config.VersionID)
	plan.CreatedAt = types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !config.UpdatedAt.IsZero() {
		plan.UpdatedAt = types.StringValue(config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		plan.UpdatedAt = types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *configResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state configResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed config value from Portkey
	config, err := r.client.GetConfig(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Portkey Config",
			"Could not read Portkey config slug "+state.Slug.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(config.ID)
	state.Slug = types.StringValue(config.Slug)
	state.Name = types.StringValue(config.Name)
	// Preserve workspace_id from state to avoid triggering RequiresReplace unnecessarily
	if state.WorkspaceID.IsNull() || state.WorkspaceID.IsUnknown() {
		state.WorkspaceID = types.StringValue(config.WorkspaceID)
	}
	state.IsDefault = types.BoolValue(config.IsDefault == 1)
	state.Status = types.StringValue(config.Status)
	state.VersionID = types.StringValue(config.VersionID)

	// Check if the API config is semantically equal to the state config
	// If so, keep the state's formatting to avoid unnecessary diffs
	apiConfigJSON := ""
	if config.Config != nil {
		if configBytes, err := json.Marshal(config.Config); err == nil {
			apiConfigJSON = string(configBytes)
		}
	} else if config.ConfigRaw != "" {
		apiConfigJSON = config.ConfigRaw
	}

	// Only update state.Config if the content differs semantically
	if !state.Config.IsNull() && !state.Config.IsUnknown() {
		// Parse both configs and compare
		var stateConfigMap, apiConfigMap map[string]interface{}
		stateErr := json.Unmarshal([]byte(state.Config.ValueString()), &stateConfigMap)
		apiErr := json.Unmarshal([]byte(apiConfigJSON), &apiConfigMap)

		// If both parse successfully and are equal, keep state's format
		if stateErr == nil && apiErr == nil {
			stateBytes, _ := json.Marshal(stateConfigMap)
			apiBytes, _ := json.Marshal(apiConfigMap)
			if string(stateBytes) != string(apiBytes) {
				// Content is different, update state
				state.Config = types.StringValue(apiConfigJSON)
			}
			// Otherwise keep existing state.Config to avoid whitespace diffs
		} else {
			state.Config = types.StringValue(apiConfigJSON)
		}
	} else {
		state.Config = types.StringValue(apiConfigJSON)
	}

	state.CreatedAt = types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !config.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan configResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for the slug
	var state configResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the config JSON string to a map
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Config.ValueString()), &configMap); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Config JSON",
			"The config attribute must be valid JSON: "+err.Error(),
		)
		return
	}

	// Update existing config
	updateReq := client.UpdateConfigRequest{
		Name:   plan.Name.ValueString(),
		Config: configMap,
	}

	_, err := r.client.UpdateConfig(ctx, state.Slug.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Portkey Config",
			"Could not update config, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated config details
	config, err := r.client.GetConfig(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading config after update",
			"Could not read config, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(config.ID)
	plan.Slug = types.StringValue(config.Slug)
	plan.WorkspaceID = types.StringValue(config.WorkspaceID)
	plan.IsDefault = types.BoolValue(config.IsDefault == 1)
	plan.Status = types.StringValue(config.Status)
	plan.VersionID = types.StringValue(config.VersionID)
	plan.CreatedAt = types.StringValue(config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !config.UpdatedAt.IsZero() {
		plan.UpdatedAt = types.StringValue(config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state configResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing config
	err := r.client.DeleteConfig(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey Config",
			"Could not delete config, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *configResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by slug
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}
