package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ datasource.DataSource              = &webhookDataSource{}
	_ datasource.DataSourceWithConfigure = &webhookDataSource{}
)

type webhookDataSource struct {
	client *client.SevallaClient
}

func NewDataSource() datasource.DataSource {
	return &webhookDataSource{}
}

func (d *webhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (d *webhookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a Sevalla webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the webhook.",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URI the webhook sends events to.",
				Computed:    true,
			},
			"allowed_events": schema.ListAttribute{
				Description: "The list of events that trigger the webhook.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the webhook is enabled.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "An optional description for the webhook.",
				Computed:    true,
			},
			"secret": schema.StringAttribute{
				Description: "The webhook signing secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"old_secret": schema.StringAttribute{
				Description: "The previous webhook signing secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"old_secret_expired_at": schema.StringAttribute{
				Description: "When the old secret expires.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The user who created the webhook.",
				Computed:    true,
			},
			"updated_by": schema.StringAttribute{
				Description: "The user who last updated the webhook.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the webhook.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the webhook was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the webhook was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *webhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *webhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model WebhookResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := d.client.GetWebhook(ctx, model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Webhook",
			"Could not read webhook ID "+model.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenWebhook(ctx, webhook, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
