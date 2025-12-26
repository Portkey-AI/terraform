package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &providerResource{}
	_ resource.ResourceWithConfigure   = &providerResource{}
	_ resource.ResourceWithImportState = &providerResource{}
)

// NewProviderResource is a helper function to simplify the provider implementation.
func NewProviderResource() resource.Resource {
	return &providerResource{}
}

// providerResource is the resource implementation.
type providerResource struct {
	client *client.Client
}

// providerResourceModel maps the resource schema data.
type providerResourceModel struct {
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

// Metadata returns the resource type name.
func (r *providerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

// Schema defines the schema for the resource.
func (r *providerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manages a Portkey Provider (also known as Virtual Key).

Providers are workspace-scoped credentials that connect to AI provider integrations. They allow you to use
organization-level integrations within specific workspaces with optional rate limits and usage controls.

**Important:** The integration_id must reference an integration that is enabled for the specified workspace.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Provider identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "URL-friendly identifier for the provider. Auto-generated if not provided.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the provider.",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "Workspace ID (UUID) where this provider will be created. Required.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"integration_id": schema.StringAttribute{
				Description: "Integration slug or ID to use. Must be an integration enabled for the workspace.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ai_provider_id": schema.StringAttribute{
				Description: "AI provider type (e.g., 'openai', 'anthropic', 'azure-openai'). Computed from the integration.",
				Computed:    true,
			},
			"note": schema.StringAttribute{
				Description: "Optional note or description for this provider.",
				Optional:    true,
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

// Configure adds the provider configured client to the resource.
func (r *providerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *providerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan providerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := client.CreateProviderRequest{
		Name:          plan.Name.ValueString(),
		WorkspaceID:   plan.WorkspaceID.ValueString(),
		IntegrationID: plan.IntegrationID.ValueString(),
	}

	if !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		createReq.Slug = plan.Slug.ValueString()
	}

	if !plan.Note.IsNull() && !plan.Note.IsUnknown() {
		createReq.Note = plan.Note.ValueString()
	}

	// Create provider
	createResp, err := r.client.CreateProvider(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating provider",
			"Could not create provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch the full provider details
	provider, err := r.client.GetProvider(ctx, createResp.ID, plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading provider after creation",
			"Could not read provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(provider.ID)
	plan.Slug = types.StringValue(provider.Slug)
	plan.Status = types.StringValue(provider.Status)
	plan.AIProviderID = types.StringValue(provider.AIProviderID)
	plan.CreatedAt = types.StringValue(provider.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if provider.Note != "" {
		plan.Note = types.StringValue(provider.Note)
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *providerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed provider value from Portkey
	provider, err := r.client.GetProvider(ctx, state.ID.ValueString(), state.WorkspaceID.ValueString())
	if err != nil {
		// Check if it's a 404 (not found)
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Portkey Provider",
			"Could not read Portkey provider ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Name = types.StringValue(provider.Name)
	state.Slug = types.StringValue(provider.Slug)
	state.Status = types.StringValue(provider.Status)
	state.AIProviderID = types.StringValue(provider.AIProviderID)
	state.CreatedAt = types.StringValue(provider.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	if provider.IntegrationID != "" {
		state.IntegrationID = types.StringValue(provider.IntegrationID)
	}

	if provider.Note != "" {
		state.Note = types.StringValue(provider.Note)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *providerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan providerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state providerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := client.UpdateProviderRequest{
		Name:        plan.Name.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
	}

	if !plan.Note.IsNull() {
		updateReq.Note = plan.Note.ValueString()
	}

	provider, err := r.client.UpdateProvider(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Portkey Provider",
			"Could not update provider, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with refreshed values
	plan.ID = types.StringValue(provider.ID)
	plan.Slug = types.StringValue(provider.Slug)
	plan.Status = types.StringValue(provider.Status)
	plan.AIProviderID = types.StringValue(provider.AIProviderID)
	plan.CreatedAt = types.StringValue(provider.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *providerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state providerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing provider
	err := r.client.DeleteProvider(ctx, state.ID.ValueString(), state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey Provider",
			"Could not delete provider, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
// Import format: "workspace_id:provider_id"
func (r *providerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: workspace_id:provider_id
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in format: workspace_id:provider_id",
		)
		return
	}

	workspaceID := parts[0]
	providerID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), providerID)...)
}

