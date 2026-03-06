package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &applicationListDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationListDataSource{}
)

type applicationListDataSource struct {
	client *client.SevallaClient
}

func NewListDataSource() datasource.DataSource {
	return &applicationListDataSource{}
}

func (d *applicationListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications"
}

func (d *applicationListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all Sevalla applications.",
		Attributes: map[string]schema.Attribute{
			"applications": schema.ListNestedAttribute{
				Description: "The list of applications.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the application.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The system-generated name of the application.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the application.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the application.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The application type.",
							Computed:    true,
						},
						"source": schema.StringAttribute{
							Description: "The source type of the application.",
							Computed:    true,
						},
						"is_suspended": schema.BoolAttribute{
							Description: "Whether the application is suspended.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the application was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "When the application was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *applicationListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationListDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apps, err := d.client.ListApplications(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Applications",
			"Could not list applications: "+err.Error(),
		)
		return
	}

	var model ApplicationListDataSourceModel
	model.Applications = make([]ApplicationListItemModel, len(apps))

	for i, app := range apps {
		model.Applications[i] = ApplicationListItemModel{
			ID:          types.StringValue(app.ID),
			Name:        types.StringValue(app.Name),
			DisplayName: types.StringValue(app.DisplayName),
			Status:      optionalString(app.Status),
			Type:        types.StringValue(app.Type),
			Source:      types.StringValue(app.Source),
			IsSuspended: types.BoolValue(app.IsSuspended),
			CreatedAt:   types.StringValue(app.CreatedAt),
			UpdatedAt:   types.StringValue(app.UpdatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
