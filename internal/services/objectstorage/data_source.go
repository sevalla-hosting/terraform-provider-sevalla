package objectstorage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &objectStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &objectStorageDataSource{}
)

type objectStorageDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &objectStorageDataSource{}
}

func (d *objectStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage"
}

func (d *objectStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla object storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the object storage.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the object storage.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the object storage.",
				Computed:    true,
			},
			"location": schema.StringAttribute{
				Description: "Geographic hint for where most data access occurs.",
				Computed:    true,
			},
			"jurisdiction": schema.StringAttribute{
				Description: "Data residency jurisdiction.",
				Computed:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The public CDN domain for accessing objects.",
				Computed:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The S3-compatible API endpoint URL.",
				Computed:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "The access key for S3-compatible API authentication.",
				Computed:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "The secret key for S3-compatible API authentication.",
				Computed:    true,
				Sensitive:   true,
			},
			"bucket_name": schema.StringAttribute{
				Description: "The bucket name.",
				Computed:    true,
			},
			"public_access": schema.BoolAttribute{
				Description: "Whether a public CDN domain is enabled for this bucket.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project this bucket is grouped under.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the object storage.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the object storage was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the object storage was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *objectStorageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *objectStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model ObjectStorageResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	os, err := d.client.GetObjectStorage(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Object Storage",
			"Could not read object storage ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenObjectStorage(ctx, os, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
