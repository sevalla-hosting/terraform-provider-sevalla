package database

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &databaseListDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseListDataSource{}
)

type databaseListDataSource struct {
	client *client.SevallaClient
}

func NewListDataSource() datasource.DataSource {
	return &databaseListDataSource{}
}

func (d *databaseListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *databaseListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all Sevalla databases.",
		Attributes: map[string]schema.Attribute{
			"databases": schema.ListNestedAttribute{
				Description: "The list of databases.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the database.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The system-generated name of the database.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the database.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The database engine type.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the database.",
							Computed:    true,
						},
						"is_suspended": schema.BoolAttribute{
							Description: "Whether the database is suspended.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the database was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "When the database was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *databaseListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *databaseListDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	dbs, err := d.client.ListDatabases(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Databases",
			"Could not list databases: "+err.Error(),
		)
		return
	}

	var model DatabaseListDataSourceModel
	model.Databases = make([]DatabaseListItemModel, len(dbs))

	for i, db := range dbs {
		model.Databases[i] = DatabaseListItemModel{
			ID:          types.StringValue(db.ID),
			Name:        types.StringValue(db.Name),
			DisplayName: types.StringValue(db.DisplayName),
			Type:        types.StringValue(db.Type),
			Status:      optionalString(db.Status),
			IsSuspended: types.BoolValue(db.IsSuspended),
			CreatedAt:   types.StringValue(db.CreatedAt),
			UpdatedAt:   types.StringValue(db.UpdatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
