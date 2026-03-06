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
	_ datasource.DataSource              = &processResourceTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &processResourceTypesDataSource{}
)

type processResourceTypesDataSource struct {
	client *client.SevallaClient
}

type ProcessResourceTypesDataSourceModel struct {
	ProcessResourceTypes []ProcessResourceTypeModel `tfsdk:"process_resource_types"`
}

type ProcessResourceTypeModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CPULimit    types.Int64  `tfsdk:"cpu_limit"`
	MemoryLimit types.Int64  `tfsdk:"memory_limit"`
	Category    types.String `tfsdk:"category"`
}

func NewProcessResourceTypesDataSource() datasource.DataSource {
	return &processResourceTypesDataSource{}
}

func (d *processResourceTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process_resource_types"
}

func (d *processResourceTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list all available Sevalla process resource types.",
		Attributes: map[string]schema.Attribute{
			"process_resource_types": schema.ListNestedAttribute{
				Description: "The list of available process resource types.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the process resource type.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the process resource type.",
							Computed:    true,
						},
						"cpu_limit": schema.Int64Attribute{
							Description: "The CPU limit in millicores.",
							Computed:    true,
						},
						"memory_limit": schema.Int64Attribute{
							Description: "The memory limit in bytes.",
							Computed:    true,
						},
						"category": schema.StringAttribute{
							Description: "The category of the process resource type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *processResourceTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *processResourceTypesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	resourceTypes, err := d.client.ListProcessResourceTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Process Resource Types",
			"Could not list process resource types: "+err.Error(),
		)
		return
	}

	var model ProcessResourceTypesDataSourceModel
	model.ProcessResourceTypes = make([]ProcessResourceTypeModel, len(resourceTypes))

	for i, rt := range resourceTypes {
		model.ProcessResourceTypes[i] = ProcessResourceTypeModel{
			ID:          types.StringValue(rt.ID),
			Name:        types.StringValue(rt.Name),
			CPULimit:    types.Int64Value(rt.CPULimit),
			MemoryLimit: types.Int64Value(rt.MemoryLimit),
			Category:    types.StringValue(rt.Category),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
