package webhook

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

type webhookResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &webhookResource{}
}

// events are the webhook event types accepted by the API.
var events = []string{
	"APP_CREATE", "APP_UPDATE", "APP_DELETE", "APP_DEPLOY",
	"DATABASE_CREATE", "DATABASE_UPDATE", "DATABASE_DELETE",
	"STATIC_SITE_CREATE", "STATIC_SITE_UPDATE", "STATIC_SITE_DELETE", "STATIC_SITE_DEPLOY",
}

var eventList = "`" + strings.Join(events, "`, `") + "`"

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": schema.StringAttribute{
				Description: "The URL that webhook events are delivered to via HTTP POST.",
				Required:    true,
			},
			"allowed_events": schema.ListAttribute{
				Description: "The events the webhook subscribes to. Valid values: " + eventList + ".",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(stringvalidator.OneOf(events...)),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the webhook delivers events. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description for the webhook (at most 255 characters).",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"secret": schema.StringAttribute{
				Description: "The webhook signing secret.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"old_secret": schema.StringAttribute{
				Description: "The previous webhook signing secret after a secret rotation.",
				Computed:    true,
				Sensitive:   true,
			},
			"old_secret_expired_at": schema.StringAttribute{
				Description: "The expiration timestamp of the old secret.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Description: "The identifier of the API key that created the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_by": schema.StringAttribute{
				Description: "The identifier of the API key that last updated the webhook.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the webhook was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the webhook was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.SevallaClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.SevallaClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.CreateWebhook(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Webhook",
			"Could not create webhook, unexpected error: "+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), webhook.ID)...)

	webhook, err = r.setEnabled(ctx, webhook, plan.IsEnabled)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Webhook", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenWebhook(ctx, webhook, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.GetWebhook(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Webhook",
			"Could not read webhook ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenWebhook(ctx, webhook, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	var state WebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateRequest(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var webhook *client.Webhook
	var err error
	if reflect.ValueOf(*updateReq).IsZero() {
		webhook, err = r.client.GetWebhook(ctx, state.ID.ValueString())
	} else {
		webhook, err = r.client.UpdateWebhook(ctx, state.ID.ValueString(), updateReq)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Webhook",
			"Could not update webhook ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	webhook, err = r.setEnabled(ctx, webhook, plan.IsEnabled)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Webhook", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenWebhook(ctx, webhook, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWebhook(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Webhook",
			"Could not delete webhook ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

// setEnabled toggles the webhook when its enabled state differs from want,
// which is skipped when want is unknown (not configured).
func (r *webhookResource) setEnabled(ctx context.Context, webhook *client.Webhook, want types.Bool) (*client.Webhook, error) {
	if want.IsUnknown() || want.IsNull() || want.ValueBool() == webhook.IsEnabled {
		return webhook, nil
	}
	toggled, err := r.client.ToggleWebhook(ctx, webhook.ID)
	if err != nil {
		return nil, fmt.Errorf("could not toggle webhook ID %s: %w", webhook.ID, err)
	}
	return toggled, nil
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
