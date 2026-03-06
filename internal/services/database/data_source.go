package database

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &databaseDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseDataSource{}
)

type databaseDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &databaseDataSource{}
}

func (d *databaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *databaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the database.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the database.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The database engine type.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The database engine version.",
				Computed:    true,
			},
			"cluster_id": schema.StringAttribute{
				Description: "The cluster where the database is deployed.",
				Computed:    true,
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The resource type (size) of the database.",
				Computed:    true,
			},
			"db_name": schema.StringAttribute{
				Description: "The database name.",
				Computed:    true,
			},
			"db_password": schema.StringAttribute{
				Description: "The database password.",
				Computed:    true,
				Sensitive:   true,
			},
			"db_user": schema.StringAttribute{
				Description: "The database user.",
				Computed:    true,
			},
			"extensions": schema.SingleNestedAttribute{
				Description: "PostgreSQL extensions.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"enable_vector": schema.BoolAttribute{
						Description: "Whether the pgvector extension is enabled.",
						Computed:    true,
					},
					"enable_postgis": schema.BoolAttribute{
						Description: "Whether the PostGIS extension is enabled.",
						Computed:    true,
					},
					"enable_cron": schema.BoolAttribute{
						Description: "Whether the pg_cron extension is enabled.",
						Computed:    true,
					},
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID associated with this database.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the database.",
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
			"cluster_display_name": schema.StringAttribute{
				Description: "The display name of the cluster.",
				Computed:    true,
			},
			"cluster_location": schema.StringAttribute{
				Description: "The location of the cluster.",
				Computed:    true,
			},
			"resource_type_name": schema.StringAttribute{
				Description: "The name of the resource type.",
				Computed:    true,
			},
			"cpu_limit": schema.Int64Attribute{
				Description: "The CPU limit for the database.",
				Computed:    true,
			},
			"memory_limit": schema.Int64Attribute{
				Description: "The memory limit for the database in bytes.",
				Computed:    true,
			},
			"storage_size": schema.Int64Attribute{
				Description: "The storage size for the database in bytes.",
				Computed:    true,
			},
			"internal_hostname": schema.StringAttribute{
				Description: "The internal hostname for connecting to the database.",
				Computed:    true,
			},
			"internal_port": schema.StringAttribute{
				Description: "The internal port for connecting to the database.",
				Computed:    true,
			},
			"external_hostname": schema.StringAttribute{
				Description: "The external hostname for connecting to the database.",
				Computed:    true,
				Sensitive:   true,
			},
			"external_port": schema.StringAttribute{
				Description: "The external port for connecting to the database.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the database.",
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
	}
}

func (d *databaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *databaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model DatabaseResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := d.client.GetDatabase(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Database",
			"Could not read database ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDatabase(ctx, db, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
