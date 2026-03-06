package staticsite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &staticSiteDataSource{}
	_ datasource.DataSourceWithConfigure = &staticSiteDataSource{}
)

type staticSiteDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &staticSiteDataSource{}
}

func (d *staticSiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site"
}

func (d *staticSiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla static site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the static site.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the static site.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source type of the static site.",
				Computed:    true,
			},
			"git_type": schema.StringAttribute{
				Description: "The Git provider type.",
				Computed:    true,
			},
			"repo_url": schema.StringAttribute{
				Description: "The repository URL for the static site source.",
				Computed:    true,
			},
			"default_branch": schema.StringAttribute{
				Description: "The default branch to deploy from.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID associated with this static site.",
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Whether auto deploy is enabled.",
				Computed:    true,
			},
			"is_preview_enabled": schema.BoolAttribute{
				Description: "Whether preview deployments are enabled.",
				Computed:    true,
			},
			"install_command": schema.StringAttribute{
				Description: "The install command.",
				Computed:    true,
			},
			"build_command": schema.StringAttribute{
				Description: "The build command.",
				Computed:    true,
			},
			"published_directory": schema.StringAttribute{
				Description: "The directory where built static files are published.",
				Computed:    true,
			},
			"root_directory": schema.StringAttribute{
				Description: "The root directory within the repository.",
				Computed:    true,
			},
			"node_version": schema.StringAttribute{
				Description: "The Node.js version used for building.",
				Computed:    true,
			},
			"index_file": schema.StringAttribute{
				Description: "The index file for the static site.",
				Computed:    true,
			},
			"error_file": schema.StringAttribute{
				Description: "The error file for the static site.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the static site.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the static site.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the static site.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the static site.",
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
	}
}

func (d *staticSiteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *staticSiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model StaticSiteResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ss, err := d.client.GetStaticSite(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Static Site",
			"Could not read static site ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenStaticSite(ctx, ss, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
