package apikey

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &apiKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &apiKeyDataSource{}
)

type apiKeyDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &apiKeyDataSource{}
}

func (d *apiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *apiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the API key.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the API key.",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "The expiration date of the API key.",
				Computed:    true,
			},
			"token": schema.StringAttribute{
				Description: "The API key token. Only available on create; will be null for existing keys.",
				Computed:    true,
				Sensitive:   true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the API key is enabled.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the API key.",
				Computed:    true,
			},
			"capabilities": schema.ListNestedAttribute{
				Description: "The capabilities (permissions) assigned to this API key.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							Description: "The permission string.",
							Computed:    true,
						},
						"id_resource": schema.StringAttribute{
							Description: "The resource ID this permission is scoped to, if any.",
							Computed:    true,
						},
					},
				},
			},
			"role_ids": schema.ListAttribute{
				Description: "The role IDs assigned to this API key.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"roles": schema.ListNestedAttribute{
				Description: "The roles assigned to this API key.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the role.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the role.",
							Computed:    true,
						},
					},
				},
			},
			"source": schema.StringAttribute{
				Description: "How the API key was created (DASHBOARD, CLI, EXTERNAL).",
				Computed:    true,
			},
			"last_used_at": schema.StringAttribute{
				Description: "When the API key was last used.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the API key was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the API key was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *apiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.SevallaClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.SevallaClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *apiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model APIKeyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := d.client.GetAPIKey(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading API Key",
			"Could not read API key ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	flattenAPIKey(apiKey, &model)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
