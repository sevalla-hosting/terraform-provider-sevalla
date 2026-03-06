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
	_ datasource.DataSource              = &apiKeyRolesDataSource{}
	_ datasource.DataSourceWithConfigure = &apiKeyRolesDataSource{}
)

type apiKeyRolesDataSource struct {
	client *client.SevallaClient
}

type APIKeyRolesDataSourceModel struct {
	Roles []APIKeyRoleModel `tfsdk:"roles"`
}

type APIKeyRoleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

func NewAPIKeyRolesDataSource() datasource.DataSource {
	return &apiKeyRolesDataSource{}
}

func (d *apiKeyRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key_roles"
}

func (d *apiKeyRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all available Sevalla API key roles.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Description: "The list of available API key roles.",
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
							Description: "Human-readable description of what this role allows.",
							Computed:    true,
						},
						"permissions": schema.ListAttribute{
							Description: "List of permission IDs granted by this role.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *apiKeyRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *apiKeyRolesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	roles, err := d.client.ListAPIKeyRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing API Key Roles",
			"Could not list API key roles: "+err.Error(),
		)
		return
	}

	var model APIKeyRolesDataSourceModel
	model.Roles = make([]APIKeyRoleModel, len(roles))

	for i, r := range roles {
		permList, diags := types.ListValueFrom(ctx, types.StringType, r.Permissions)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.Roles[i] = APIKeyRoleModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
			Permissions: permList,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
