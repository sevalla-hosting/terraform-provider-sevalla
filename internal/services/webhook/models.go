package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// WebhookResourceModel is the Terraform state model for the sevalla_webhook resource.
type WebhookResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Endpoint           types.String `tfsdk:"endpoint"`
	AllowedEvents      types.List   `tfsdk:"allowed_events"`
	IsEnabled          types.Bool   `tfsdk:"is_enabled"`
	Secret             types.String `tfsdk:"secret"`
	OldSecret          types.String `tfsdk:"old_secret"`
	OldSecretExpiredAt types.String `tfsdk:"old_secret_expired_at"`
	Description        types.String `tfsdk:"description"`
	CompanyID          types.String `tfsdk:"company_id"`
	CreatedBy          types.String `tfsdk:"created_by"`
	UpdatedBy          types.String `tfsdk:"updated_by"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

// flattenWebhook maps a client.Webhook to the Terraform resource model.
func flattenWebhook(ctx context.Context, w *client.Webhook, model *WebhookResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(w.ID)
	model.Endpoint = types.StringValue(w.Endpoint)
	model.IsEnabled = types.BoolValue(w.IsEnabled)
	model.Secret = optionalString(w.Secret)
	model.OldSecret = optionalString(w.OldSecret)
	model.OldSecretExpiredAt = optionalString(w.OldSecretExpiredAt)
	model.Description = optionalString(w.Description)
	model.CompanyID = optionalString(w.CompanyID)
	model.CreatedBy = optionalString(w.CreatedBy)
	model.UpdatedBy = optionalString(w.UpdatedBy)
	model.CreatedAt = types.StringValue(w.CreatedAt)
	model.UpdatedAt = types.StringValue(w.UpdatedAt)

	allowedEvents, d := types.ListValueFrom(ctx, types.StringType, w.AllowedEvents)
	diags.Append(d...)
	model.AllowedEvents = allowedEvents

	return diags
}

// buildCreateRequest constructs a CreateWebhookRequest from the Terraform plan model.
func buildCreateRequest(ctx context.Context, model *WebhookResourceModel) (*client.CreateWebhookRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var allowedEvents []string
	diags.Append(model.AllowedEvents.ElementsAs(ctx, &allowedEvents, false)...)
	if diags.HasError() {
		return nil, diags
	}

	req := &client.CreateWebhookRequest{
		AllowedEvents: allowedEvents,
		Endpoint:      model.Endpoint.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		v := model.Description.ValueString()
		req.Description = &v
	}

	return req, diags
}

// buildUpdateRequest constructs an UpdateWebhookRequest from the Terraform plan model.
func buildUpdateRequest(ctx context.Context, plan *WebhookResourceModel, state *WebhookResourceModel) (*client.UpdateWebhookRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateWebhookRequest{}

	if !plan.Endpoint.Equal(state.Endpoint) {
		v := plan.Endpoint.ValueString()
		req.Endpoint = &v
	}
	if !plan.AllowedEvents.Equal(state.AllowedEvents) {
		var allowedEvents []string
		diags.Append(plan.AllowedEvents.ElementsAs(ctx, &allowedEvents, false)...)
		req.AllowedEvents = allowedEvents
	}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			req.Description = nil
		} else {
			v := plan.Description.ValueString()
			req.Description = &v
		}
	}

	return req, diags
}

// optionalString converts a *string to a types.String, returning null for nil pointers.
func optionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
