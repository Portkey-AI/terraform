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
	_ datasource.DataSource              = &promptsDataSource{}
	_ datasource.DataSourceWithConfigure = &promptsDataSource{}
)

// NewPromptsDataSource is a helper function to simplify the provider implementation.
func NewPromptsDataSource() datasource.DataSource {
	return &promptsDataSource{}
}

// promptsDataSource is the data source implementation.
type promptsDataSource struct {
	client *client.Client
}

// promptsDataSourceModel maps the data source schema data.
type promptsDataSourceModel struct {
	WorkspaceID  types.String         `tfsdk:"workspace_id"`
	CollectionID types.String         `tfsdk:"collection_id"`
	Prompts      []promptSummaryModel `tfsdk:"prompts"`
}

// promptSummaryModel maps prompt summary data.
type promptSummaryModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	CollectionID types.String `tfsdk:"collection_id"`
	Model        types.String `tfsdk:"model"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *promptsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompts"
}

// Schema defines the schema for the data source.
func (d *promptsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get a list of Portkey prompts.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "Optional workspace ID to filter prompts.",
				Optional:    true,
			},
			"collection_id": schema.StringAttribute{
				Description: "Optional collection ID to filter prompts.",
				Optional:    true,
			},
			"prompts": schema.ListNestedAttribute{
				Description: "List of prompts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Prompt identifier (UUID).",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "URL-friendly identifier for the prompt.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Human-readable name for the prompt.",
							Computed:    true,
						},
						"collection_id": schema.StringAttribute{
							Description: "Collection ID the prompt belongs to.",
							Computed:    true,
						},
						"model": schema.StringAttribute{
							Description: "Model used for this prompt.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the prompt (active, archived).",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the prompt was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Timestamp when the prompt was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *promptsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *promptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state promptsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := ""
	if !state.WorkspaceID.IsNull() {
		workspaceID = state.WorkspaceID.ValueString()
	}

	collectionID := ""
	if !state.CollectionID.IsNull() {
		collectionID = state.CollectionID.ValueString()
	}

	prompts, err := d.client.ListPrompts(ctx, workspaceID, collectionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Prompts",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, prompt := range prompts {
		promptState := promptSummaryModel{
			ID:           types.StringValue(prompt.ID),
			Slug:         types.StringValue(prompt.Slug),
			Name:         types.StringValue(prompt.Name),
			CollectionID: types.StringValue(prompt.CollectionID),
			Model:        types.StringValue(prompt.Model),
			Status:       types.StringValue(prompt.Status),
			CreatedAt:    types.StringValue(prompt.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}

		if !prompt.UpdatedAt.IsZero() {
			promptState.UpdatedAt = types.StringValue(prompt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
		}

		state.Prompts = append(state.Prompts, promptState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
