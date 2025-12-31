package provider

import (
	"context"
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
	_ resource.Resource                = &integrationResource{}
	_ resource.ResourceWithConfigure   = &integrationResource{}
	_ resource.ResourceWithImportState = &integrationResource{}
)

// NewIntegrationResource is a helper function to simplify the provider implementation.
func NewIntegrationResource() resource.Resource {
	return &integrationResource{}
}

// integrationResource is the resource implementation.
type integrationResource struct {
	client *client.Client
}

// integrationResourceModel maps the resource schema data.
type integrationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	AIProviderID types.String `tfsdk:"ai_provider_id"`
	Key          types.String `tfsdk:"key"`
	Description  types.String `tfsdk:"description"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *integrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

// Schema defines the schema for the resource.
func (r *integrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Portkey integration. Integrations connect Portkey to AI providers like OpenAI, Anthropic, Azure, etc.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Integration identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the integration. Auto-generated if not provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the integration.",
				Required:    true,
			},
			"ai_provider_id": schema.StringAttribute{
				Description: "ID of the AI provider (e.g., 'openai', 'anthropic', 'azure-openai').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "API key for the provider. This is write-only and will not be returned by the API.",
				Optional:    true,
				Sensitive:   true,
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the integration.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the integration (active, archived).",
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
	}
}

// Configure adds the provider configured client to the resource.
func (r *integrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan integrationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new integration
	createReq := client.CreateIntegrationRequest{
		Name:         plan.Name.ValueString(),
		AIProviderID: plan.AIProviderID.ValueString(),
	}

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		createReq.Slug = plan.Slug.ValueString()
	}

	if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
		createReq.Key = plan.Key.ValueString()
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = plan.Description.ValueString()
	}

	createResp, err := r.client.CreateIntegration(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating integration",
			"Could not create integration, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the full integration details
	integration, err := r.client.GetIntegration(ctx, createResp.Slug)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading integration after creation",
			"Could not read integration, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema
	plan.ID = types.StringValue(integration.ID)
	plan.Slug = types.StringValue(integration.Slug)
	plan.Status = types.StringValue(integration.Status)
	plan.CreatedAt = types.StringValue(integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !integration.UpdatedAt.IsZero() {
		plan.UpdatedAt = types.StringValue(integration.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		plan.UpdatedAt = types.StringValue(integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state integrationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed integration value from Portkey
	integration, err := r.client.GetIntegration(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Portkey Integration",
			"Could not read Portkey integration slug "+state.Slug.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(integration.ID)
	// Preserve slug from state to avoid triggering RequiresReplace unnecessarily
	if state.Slug.IsNull() || state.Slug.IsUnknown() {
		state.Slug = types.StringValue(integration.Slug)
	}
	state.Name = types.StringValue(integration.Name)
	// Preserve ai_provider_id from state to avoid triggering RequiresReplace unnecessarily
	if state.AIProviderID.IsNull() || state.AIProviderID.IsUnknown() {
		state.AIProviderID = types.StringValue(integration.AIProviderID)
	}
	state.Status = types.StringValue(integration.Status)
	if integration.Description != "" {
		state.Description = types.StringValue(integration.Description)
	}
	state.CreatedAt = types.StringValue(integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !integration.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(integration.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan integrationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for the slug
	var state integrationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing integration
	updateReq := client.UpdateIntegrationRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
		updateReq.Key = plan.Key.ValueString()
	}

	if !plan.Description.IsNull() {
		updateReq.Description = plan.Description.ValueString()
	}

	integration, err := r.client.UpdateIntegration(ctx, state.Slug.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Portkey Integration",
			"Could not update integration, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(integration.ID)
	plan.Slug = types.StringValue(integration.Slug)
	plan.Status = types.StringValue(integration.Status)
	plan.CreatedAt = types.StringValue(integration.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !integration.UpdatedAt.IsZero() {
		plan.UpdatedAt = types.StringValue(integration.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state integrationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing integration
	err := r.client.DeleteIntegration(ctx, state.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey Integration",
			"Could not delete integration, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by slug
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}
