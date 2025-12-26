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
	_ datasource.DataSource              = &promptDataSource{}
	_ datasource.DataSourceWithConfigure = &promptDataSource{}
)

// NewPromptDataSource is a helper function to simplify the provider implementation.
func NewPromptDataSource() datasource.DataSource {
	return &promptDataSource{}
}

// promptDataSource is the data source implementation.
type promptDataSource struct {
	client *client.Client
}

// promptDataSourceModel maps the data source schema data.
type promptDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Slug                types.String `tfsdk:"slug"`
	Name                types.String `tfsdk:"name"`
	CollectionID        types.String `tfsdk:"collection_id"`
	Template            types.String `tfsdk:"template"`
	Parameters          types.String `tfsdk:"parameters"`
	Model               types.String `tfsdk:"model"`
	VirtualKey          types.String `tfsdk:"virtual_key"`
	Version             types.String `tfsdk:"version"`
	PromptVersion       types.Int64  `tfsdk:"prompt_version"`
	PromptVersionID     types.String `tfsdk:"prompt_version_id"`
	PromptVersionStatus types.String `tfsdk:"prompt_version_status"`
	Status              types.String `tfsdk:"status"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

// Metadata returns the data source type name.
func (d *promptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prompt"
}

// Schema defines the schema for the data source.
func (d *promptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Portkey prompt.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The slug of the prompt to look up.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version to retrieve: 'latest', 'default', or a specific version number. Defaults to 'default'.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "Prompt identifier (UUID).",
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
			"template": schema.StringAttribute{
				Description: "Prompt template string.",
				Computed:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "JSON string of model parameters.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				Description: "Model used for this prompt.",
				Computed:    true,
			},
			"virtual_key": schema.StringAttribute{
				Description: "Virtual key (provider) slug used for this prompt.",
				Computed:    true,
			},
			"prompt_version": schema.Int64Attribute{
				Description: "Version number of the prompt.",
				Computed:    true,
			},
			"prompt_version_id": schema.StringAttribute{
				Description: "Version ID of the prompt.",
				Computed:    true,
			},
			"prompt_version_status": schema.StringAttribute{
				Description: "Status of the version (active, archived).",
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *promptDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *promptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state promptDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	version := ""
	if !state.Version.IsNull() {
		version = state.Version.ValueString()
	}

	prompt, err := d.client.GetPrompt(ctx, state.Slug.ValueString(), version)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Portkey Prompt",
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.StringValue(prompt.ID)
	state.Slug = types.StringValue(prompt.Slug)
	state.Name = types.StringValue(prompt.Name)
	state.CollectionID = types.StringValue(prompt.CollectionID)
	state.Template = types.StringValue(prompt.String)
	state.Model = types.StringValue(prompt.Model)
	state.VirtualKey = types.StringValue(prompt.VirtualKey)
	state.PromptVersion = types.Int64Value(int64(prompt.PromptVersion))
	state.PromptVersionID = types.StringValue(prompt.PromptVersionID)
	state.PromptVersionStatus = types.StringValue(prompt.PromptVersionStatus)
	state.Status = types.StringValue(prompt.Status)

	// Convert parameters to JSON string
	if prompt.Parameters != nil {
		paramsBytes, err := json.Marshal(prompt.Parameters)
		if err == nil {
			state.Parameters = types.StringValue(string(paramsBytes))
		}
	}

	state.CreatedAt = types.StringValue(prompt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	if !prompt.UpdatedAt.IsZero() {
		state.UpdatedAt = types.StringValue(prompt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
