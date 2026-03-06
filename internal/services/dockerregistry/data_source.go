package dockerregistry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &dockerRegistryDataSource{}
	_ datasource.DataSourceWithConfigure = &dockerRegistryDataSource{}
)

type dockerRegistryDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &dockerRegistryDataSource{}
}

func (d *dockerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_registry"
}

func (d *dockerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla Docker registry credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Docker registry credential.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Docker registry credential.",
				Computed:    true,
			},
			"registry": schema.StringAttribute{
				Description: "The Docker registry type (e.g., gcr, ecr, dockerHub, github, gitlab, digitalOcean, custom).",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for the Docker registry.",
				Computed:    true,
				Sensitive:   true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the Docker registry credential.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the Docker registry credential was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the Docker registry credential was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *dockerRegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// DockerRegistryDataSourceModel is a read-only model for the data source (no secret field).
type DockerRegistryDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Registry  types.String `tfsdk:"registry"`
	Username  types.String `tfsdk:"username"`
	CompanyID types.String `tfsdk:"company_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *dockerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model DockerRegistryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := d.client.GetDockerRegistry(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Docker Registry",
			"Could not read Docker registry ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	model.ID = types.StringValue(registry.ID)
	model.Name = types.StringValue(registry.Name)
	model.Registry = optionalString(registry.Registry)
	model.Username = optionalString(registry.Username)
	model.CompanyID = optionalString(registry.CompanyID)
	model.CreatedAt = types.StringValue(registry.CreatedAt)
	model.UpdatedAt = types.StringValue(registry.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
