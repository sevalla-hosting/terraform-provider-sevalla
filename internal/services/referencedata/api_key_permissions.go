package referencedata

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &apiKeyPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &apiKeyPermissionsDataSource{}
)

type apiKeyPermissionsDataSource struct {
	client *client.SevallaClient
}

type APIKeyPermissionsDataSourceModel struct {
	Permissions []APIKeyPermissionModel `tfsdk:"permissions"`
}

type APIKeyPermissionModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Resource    types.String `tfsdk:"resource"`
	Action      types.String `tfsdk:"action"`
}

func NewAPIKeyPermissionsDataSource() datasource.DataSource {
	return &apiKeyPermissionsDataSource{}
}

func (d *apiKeyPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key_permissions"
}

func (d *apiKeyPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all available Sevalla API key permissions.",
		Attributes: map[string]schema.Attribute{
			"permissions": schema.ListNestedAttribute{
				Description: "The list of available API key permissions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the permission.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the permission (e.g., app:read).",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Human-readable description of what this permission allows.",
							Computed:    true,
						},
						"resource": schema.StringAttribute{
							Description: "The resource type this permission applies to (e.g., APP, DATABASE, PIPELINE).",
							Computed:    true,
						},
						"action": schema.StringAttribute{
							Description: "The action this permission allows (CREATE, READ, UPDATE, DELETE).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *apiKeyPermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *apiKeyPermissionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	permissions, err := d.client.ListAPIKeyPermissions(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing API Key Permissions",
			"Could not list API key permissions: "+err.Error(),
		)
		return
	}

	var model APIKeyPermissionsDataSourceModel
	model.Permissions = make([]APIKeyPermissionModel, len(permissions))

	for i, p := range permissions {
		model.Permissions[i] = APIKeyPermissionModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			Resource:    types.StringValue(p.Resource),
			Action:      types.StringValue(p.Action),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
