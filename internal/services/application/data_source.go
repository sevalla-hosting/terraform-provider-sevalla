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
	_ datasource.DataSource              = &applicationDataSource{}
	_ datasource.DataSourceWithConfigure = &applicationDataSource{}
)

type applicationDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

func (d *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the application.",
				Computed:    true,
			},
			"cluster_id": schema.StringAttribute{
				Description: "The cluster where the application is deployed.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source type of the application.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID associated with this application.",
				Computed:    true,
			},
			"git_type": schema.StringAttribute{
				Description: "The Git provider type.",
				Computed:    true,
			},
			"repo_url": schema.StringAttribute{
				Description: "The repository URL for the application source.",
				Computed:    true,
			},
			"default_branch": schema.StringAttribute{
				Description: "The default branch to deploy from.",
				Computed:    true,
			},
			"docker_image": schema.StringAttribute{
				Description: "The Docker image to deploy.",
				Computed:    true,
			},
			"docker_registry_credential_id": schema.StringAttribute{
				Description: "The ID of the Docker registry credential.",
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Whether auto deploy is enabled.",
				Computed:    true,
			},
			"build_type": schema.StringAttribute{
				Description: "The build type.",
				Computed:    true,
			},
			"build_path": schema.StringAttribute{
				Description: "The build path within the repository.",
				Computed:    true,
			},
			"build_cache_enabled": schema.BoolAttribute{
				Description: "Whether the build cache is enabled.",
				Computed:    true,
			},
			"hibernation_enabled": schema.BoolAttribute{
				Description: "Whether hibernation is enabled.",
				Computed:    true,
			},
			"hibernate_after_seconds": schema.Int64Attribute{
				Description: "Seconds of inactivity before hibernation.",
				Computed:    true,
			},
			"dockerfile_path": schema.StringAttribute{
				Description: "The path to the Dockerfile.",
				Computed:    true,
			},
			"docker_context": schema.StringAttribute{
				Description: "The Docker build context path.",
				Computed:    true,
			},
			"pack_builder": schema.StringAttribute{
				Description: "The Cloud Native Buildpack builder.",
				Computed:    true,
			},
			"nixpacks_version": schema.StringAttribute{
				Description: "The Nixpacks version.",
				Computed:    true,
			},
			"allow_deploy_paths": schema.ListAttribute{
				Description: "Paths that trigger a deployment when changed.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"ignore_deploy_paths": schema.ListAttribute{
				Description: "Paths to ignore for deployment triggers.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"buildpacks": schema.ListNestedAttribute{
				Description: "List of buildpack configurations for the application.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"order": schema.Int64Attribute{
							Description: "The order of the buildpack.",
							Computed:    true,
						},
						"source": schema.StringAttribute{
							Description: "The source of the buildpack.",
							Computed:    true,
						},
					},
				},
			},
			"wait_for_checks": schema.BoolAttribute{
				Description: "Whether to wait for checks before deploying.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the application.",
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The Kubernetes namespace.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the application.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The application type.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the application.",
				Computed:    true,
			},
			"is_suspended": schema.BoolAttribute{
				Description: "Whether the application is suspended.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The user who created the application.",
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
	}
}

func (d *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model ApplicationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplication(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenApplication(ctx, app, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
