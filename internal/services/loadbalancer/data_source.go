package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &loadBalancerDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancerDataSource{}
)

type loadBalancerDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &loadBalancerDataSource{}
}

func (d *loadBalancerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (d *loadBalancerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla load balancer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the load balancer.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the load balancer.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The routing type. Valid values: DEFAULT, GEO.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID associated with this load balancer.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the load balancer.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the load balancer.",
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
	}
}

func (d *loadBalancerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loadBalancerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model LoadBalancerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := d.client.GetLoadBalancer(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancer",
			"Could not read load balancer ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenLoadBalancer(ctx, lb, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
