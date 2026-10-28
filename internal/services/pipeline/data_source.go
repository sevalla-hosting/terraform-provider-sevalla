package pipeline

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sevalla/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &pipelineDataSource{}
	_ datasource.DataSourceWithConfigure = &pipelineDataSource{}
)

type pipelineDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &pipelineDataSource{}
}

func (d *pipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (d *pipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla pipeline.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the pipeline.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the pipeline.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the pipeline (trunk or branch).",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID associated with the pipeline.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the pipeline.",
				Computed:    true,
			},
			"stages": schema.ListNestedAttribute{
				Description: "The stages of the pipeline ordered by position, including the Dev and Production stages the API creates with the pipeline.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the stage.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "The display name of the stage.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The stage type: `standard`, `prod` or `preview`.",
							Computed:    true,
						},
						"order": schema.Int64Attribute{
							Description: "The position of the stage, starting from 0.",
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "When the pipeline was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the pipeline was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *pipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model PipelineResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipeline, err := d.client.GetPipeline(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Pipeline",
			"Could not read pipeline ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenPipeline(ctx, pipeline, &model)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
