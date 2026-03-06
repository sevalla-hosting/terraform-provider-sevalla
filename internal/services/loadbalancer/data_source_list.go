package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &loadBalancerListDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerListDataSource{}
)

type loadBalancerListDataSource struct {
	client *client.SevallaClient
}

func NewListDataSource() datasource.DataSource {
	return &loadBalancerListDataSource{}
}

func (d *loadBalancerListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

func (d *loadBalancerListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all Sevalla load balancers.",
		Attributes: map[string]schema.Attribute{
			"load_balancers": schema.ListNestedAttribute{
				Description: "The list of load balancers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the load balancer.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The system-generated name of the load balancer.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the load balancer.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The routing type.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "When the load balancer was created.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "When the load balancer was last updated.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *loadBalancerListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerListDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	lbs, err := d.client.ListLoadBalancers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Load Balancers",
			"Could not list load balancers: "+err.Error(),
		)
		return
	}

	var model LoadBalancerListDataSourceModel
	model.LoadBalancers = make([]LoadBalancerListItemModel, len(lbs))

	for i, lb := range lbs {
		model.LoadBalancers[i] = LoadBalancerListItemModel{
			ID:          types.StringValue(lb.ID),
			Name:        types.StringValue(lb.Name),
			DisplayName: types.StringValue(lb.DisplayName),
			Type:        optionalString(lb.Type),
			CreatedAt:   types.StringValue(lb.CreatedAt),
			UpdatedAt:   types.StringValue(lb.UpdatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
