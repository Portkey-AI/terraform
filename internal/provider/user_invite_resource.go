package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-portkey/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userInviteResource{}
	_ resource.ResourceWithConfigure   = &userInviteResource{}
	_ resource.ResourceWithImportState = &userInviteResource{}
)

// NewUserInviteResource is a helper function to simplify the provider implementation.
func NewUserInviteResource() resource.Resource {
	return &userInviteResource{}
}

// userInviteResource is the resource implementation.
type userInviteResource struct {
	client *client.Client
}

// userInviteResourceModel maps the resource schema data.
type userInviteResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	Workspaces types.List   `tfsdk:"workspaces"`
	Scopes     types.List   `tfsdk:"scopes"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
}

// workspaceInviteModel maps workspace invite details
type workspaceInviteModel struct {
	ID   types.String `tfsdk:"id"`
	Role types.String `tfsdk:"role"`
}

// Metadata returns the resource type name.
func (r *userInviteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_invite"
}

// Schema defines the schema for the resource.
func (r *userInviteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Portkey user invitation. Sends invitations to users to join the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User invite identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user to invite.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "Organization role for the user (e.g., 'admin', 'member').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the invitation (e.g., 'pending', 'accepted', 'expired').",
				Computed:    true,
			},
			"workspaces": schema.ListNestedAttribute{
				Description: "List of workspaces to add the user to with specific roles.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Workspace ID.",
							Required:    true,
						},
						"role": schema.StringAttribute{
							Description: "Role in the workspace (e.g., 'admin', 'member', 'manager').",
							Required:    true,
						},
					},
				},
			},
			"scopes": schema.ListAttribute{
				Description: "List of API scopes to grant to the user's workspace API key.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the invitation was created.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Timestamp when the invitation expires.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *userInviteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *userInviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userInviteResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build invite request
	inviteReq := client.CreateUserInviteRequest{
		Email: plan.Email.ValueString(),
		Role:  plan.Role.ValueString(),
	}

	// Add workspaces if specified
	if !plan.Workspaces.IsNull() && !plan.Workspaces.IsUnknown() {
		var workspaces []workspaceInviteModel
		diags = plan.Workspaces.ElementsAs(ctx, &workspaces, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, ws := range workspaces {
			inviteReq.Workspaces = append(inviteReq.Workspaces, client.WorkspaceInviteDetails{
				ID:   ws.ID.ValueString(),
				Role: ws.Role.ValueString(),
			})
		}
	}

	// Add scopes if specified
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var scopes []string
		diags = plan.Scopes.ElementsAs(ctx, &scopes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(scopes) > 0 {
			inviteReq.WorkspaceAPIKeyDetails = &client.APIKeyDetails{
				Scopes: scopes,
			}
		}
	}

	// Send invitation
	invite, err := r.client.InviteUser(ctx, inviteReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user invitation",
			"Could not invite user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(invite.ID)
	plan.Status = types.StringValue(invite.Status)
	plan.CreatedAt = types.StringValue(invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	plan.ExpiresAt = types.StringValue(invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"))

	// Map workspaces from response
	if len(invite.Workspaces) > 0 {
		workspaceElements := make([]attr.Value, 0, len(invite.Workspaces))
		for _, ws := range invite.Workspaces {
			workspaceElements = append(workspaceElements, types.ObjectValueMust(
				map[string]attr.Type{
					"id":   types.StringType,
					"role": types.StringType,
				},
				map[string]attr.Value{
					"id":   types.StringValue(ws.ID),
					"role": types.StringValue(ws.Role),
				},
			))
		}
		plan.Workspaces = types.ListValueMust(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"role": types.StringType,
				},
			},
			workspaceElements,
		)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userInviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userInviteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed invite value from Portkey
	invite, err := r.client.GetUserInvite(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Portkey User Invite",
			"Could not read Portkey user invite ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Email = types.StringValue(invite.Email)
	state.Role = types.StringValue(invite.Role)
	state.Status = types.StringValue(invite.Status)
	state.CreatedAt = types.StringValue(invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	state.ExpiresAt = types.StringValue(invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"))

	// Map workspaces from response
	if len(invite.Workspaces) > 0 {
		workspaceElements := make([]attr.Value, 0, len(invite.Workspaces))
		for _, ws := range invite.Workspaces {
			workspaceElements = append(workspaceElements, types.ObjectValueMust(
				map[string]attr.Type{
					"id":   types.StringType,
					"role": types.StringType,
				},
				map[string]attr.Value{
					"id":   types.StringValue(ws.ID),
					"role": types.StringValue(ws.Role),
				},
			))
		}
		state.Workspaces = types.ListValueMust(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":   types.StringType,
					"role": types.StringType,
				},
			},
			workspaceElements,
		)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userInviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// User invites cannot be updated - they must be deleted and recreated
	resp.Diagnostics.AddError(
		"User Invite Update Not Supported",
		"User invitations cannot be updated. Please delete and recreate the invitation with the new configuration.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userInviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state userInviteResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user invitation
	err := r.client.DeleteUserInvite(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Portkey User Invite",
			"Could not delete user invitation, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *userInviteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
