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
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

type usersDataSource struct {
	client *client.SevallaClient
}

type UsersDataSourceModel struct {
	Users []UserModel `tfsdk:"users"`
}

type UserModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	FullName types.String `tfsdk:"full_name"`
	Image    types.String `tfsdk:"image"`
}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all Sevalla company users.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "The list of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the user.",
							Computed:    true,
						},
						"full_name": schema.StringAttribute{
							Description: "The full name of the user.",
							Computed:    true,
						},
						"image": schema.StringAttribute{
							Description: "The avatar image URL of the user.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Users",
			"Could not list users: "+err.Error(),
		)
		return
	}

	var model UsersDataSourceModel
	model.Users = make([]UserModel, len(users))

	for i, user := range users {
		model.Users[i] = UserModel{
			ID:       types.StringValue(user.ID),
			Email:    types.StringValue(user.Email),
			FullName: types.StringValue(user.FullName),
			Image:    types.StringValue(user.Image),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
