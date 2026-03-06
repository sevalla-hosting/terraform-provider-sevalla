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
	_ datasource.DataSource              = &clustersDataSource{}
	_ datasource.DataSourceWithConfigure = &clustersDataSource{}
)

type clustersDataSource struct {
	client *client.SevallaClient
}

type ClustersDataSourceModel struct {
	Clusters []ClusterModel `tfsdk:"clusters"`
}

type ClusterModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Location    types.String `tfsdk:"location"`
}

func NewClustersDataSource() datasource.DataSource {
	return &clustersDataSource{}
}

func (d *clustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *clustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all available Sevalla clusters.",
		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				Description: "The list of available clusters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the cluster.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The system name of the cluster.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the cluster.",
							Computed:    true,
						},
						"location": schema.StringAttribute{
							Description: "The geographic location of the cluster.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *clustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clustersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	clusters, err := d.client.ListClusters(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Clusters",
			"Could not list clusters: "+err.Error(),
		)
		return
	}

	var model ClustersDataSourceModel
	model.Clusters = make([]ClusterModel, len(clusters))

	for i, cluster := range clusters {
		var displayName types.String
		if cluster.DisplayName != nil {
			displayName = types.StringValue(*cluster.DisplayName)
		} else {
			displayName = types.StringNull()
		}

		model.Clusters[i] = ClusterModel{
			ID:          types.StringValue(cluster.ID),
			Name:        types.StringValue(cluster.Name),
			DisplayName: displayName,
			Location:    types.StringValue(cluster.Location),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
