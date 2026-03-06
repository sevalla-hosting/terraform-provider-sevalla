package staticsite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &staticSiteListDataSource{}
	_ datasource.DataSourceWithConfigure = &staticSiteListDataSource{}
)

type staticSiteListDataSource struct {
	client *client.SevallaClient
}

func NewListDataSource() datasource.DataSource {
	return &staticSiteListDataSource{}
}

func (d *staticSiteListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_sites"
}

func (d *staticSiteListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all Sevalla static sites.",
		Attributes: map[string]schema.Attribute{
			"static_sites": schema.ListNestedAttribute{
				Description: "The list of static sites.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the static site.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The system-generated name of the static site.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the static site.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the static site.",
							Computed:    true,
						},
						"source": schema.StringAttribute{
							Description: "The source type of the static site.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the static site was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "When the static site was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *staticSiteListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *staticSiteListDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	sites, err := d.client.ListStaticSites(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Static Sites",
			"Could not list static sites: "+err.Error(),
		)
		return
	}

	var model StaticSiteListDataSourceModel
	model.StaticSites = make([]StaticSiteListItemModel, len(sites))

	for i, site := range sites {
		model.StaticSites[i] = StaticSiteListItemModel{
			ID:          types.StringValue(site.ID),
			Name:        types.StringValue(site.Name),
			DisplayName: types.StringValue(site.DisplayName),
			Status:      optionalString(site.Status),
			Source:      types.StringValue(site.Source),
			CreatedAt:   types.StringValue(site.CreatedAt),
			UpdatedAt:   types.StringValue(site.UpdatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
