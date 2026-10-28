package apikey

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// APIKeyResourceModel is the Terraform state model for the sevalla_api_key resource.
type APIKeyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
	Token        types.String `tfsdk:"token"`
	MaskedToken  types.String `tfsdk:"masked_token"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	CompanyID    types.String `tfsdk:"company_id"`
	Capabilities types.List   `tfsdk:"capabilities"`
	RoleIDs      types.List   `tfsdk:"role_ids"`
	Roles        types.List   `tfsdk:"roles"`
	Source       types.String `tfsdk:"source"`
	LastUsedAt   types.String `tfsdk:"last_used_at"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// APIKeyCapabilityModel represents a single capability with optional resource scoping.
type APIKeyCapabilityModel struct {
	Permission types.String `tfsdk:"permission"`
	IDResource types.String `tfsdk:"id_resource"`
}

// APIKeyRoleModel represents a role assigned to an API key (computed from API response).
type APIKeyRoleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// capabilityObjectType is the attr.Type for APIKeyCapabilityModel.
var capabilityObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"permission":  types.StringType,
		"id_resource": types.StringType,
	},
}

// roleObjectType is the attr.Type for APIKeyRoleModel.
var roleObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
	},
}

// flattenAPIKey maps a client.APIKey to the Terraform model. The token is only
// returned on create, so the model's existing token is left untouched.
func flattenAPIKey(ctx context.Context, k *client.APIKey, model *APIKeyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(k.ID)
	model.Name = types.StringValue(k.Name)
	model.MaskedToken = optionalString(k.MaskedToken)
	model.Enabled = types.BoolValue(k.Enabled)
	model.CompanyID = optionalString(k.CompanyID)
	model.Source = optionalString(k.Source)
	model.LastUsedAt = optionalString(k.LastUsedAt)
	model.CreatedAt = types.StringValue(k.CreatedAt)
	model.UpdatedAt = types.StringValue(k.UpdatedAt)

	// Keep the configured timestamp when the API returns the same instant in
	// another format (e.g. with milliseconds).
	if k.ExpiresAt == nil || !sameInstant(model.ExpiresAt.ValueString(), *k.ExpiresAt) {
		model.ExpiresAt = optionalString(k.ExpiresAt)
	}

	var priorCaps []APIKeyCapabilityModel
	if !model.Capabilities.IsNull() && !model.Capabilities.IsUnknown() {
		diags.Append(model.Capabilities.ElementsAs(ctx, &priorCaps, false)...)
	}
	caps := make([]APIKeyCapabilityModel, 0, len(k.Capabilities))
	for _, c := range k.Capabilities {
		caps = append(caps, APIKeyCapabilityModel{
			Permission: types.StringValue(c.Permission),
			IDResource: optionalString(c.ResourceID),
		})
	}
	if len(caps) == 0 {
		model.Capabilities = types.ListNull(capabilityObjectType)
	} else {
		v, d := types.ListValueFrom(ctx, capabilityObjectType, keepOrder(priorCaps, caps))
		diags.Append(d...)
		model.Capabilities = v
	}

	var priorRoleIDs []string
	if !model.RoleIDs.IsNull() && !model.RoleIDs.IsUnknown() {
		diags.Append(model.RoleIDs.ElementsAs(ctx, &priorRoleIDs, false)...)
	}
	roleIDs := make([]string, 0, len(k.Roles))
	roles := make([]APIKeyRoleModel, 0, len(k.Roles))
	for _, r := range k.Roles {
		roleIDs = append(roleIDs, r.ID)
		roles = append(roles, APIKeyRoleModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
		})
	}
	if len(roleIDs) == 0 {
		model.RoleIDs = types.ListNull(types.StringType)
	} else {
		v, d := types.ListValueFrom(ctx, types.StringType, keepOrder(priorRoleIDs, roleIDs))
		diags.Append(d...)
		model.RoleIDs = v
	}
	roleList, d := types.ListValueFrom(ctx, roleObjectType, roles)
	diags.Append(d...)
	model.Roles = roleList

	return diags
}

func buildCapabilities(ctx context.Context, list types.List) ([]client.APIKeyCapability, diag.Diagnostics) {
	// A non-nil empty slice clears the capabilities on update.
	out := []client.APIKeyCapability{}
	if list.IsNull() || list.IsUnknown() {
		return out, nil
	}
	var caps []APIKeyCapabilityModel
	diags := list.ElementsAs(ctx, &caps, false)
	for _, c := range caps {
		out = append(out, client.APIKeyCapability{
			Permission: c.Permission.ValueString(),
			ResourceID: c.IDResource.ValueStringPointer(),
		})
	}
	return out, diags
}

func buildRoleIDs(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	// A non-nil empty slice clears the roles on update.
	out := []string{}
	if list.IsNull() || list.IsUnknown() {
		return out, nil
	}
	diags := list.ElementsAs(ctx, &out, false)
	return out, diags
}

// buildCreateRequest constructs a CreateAPIKeyRequest from the Terraform plan model.
func buildCreateRequest(ctx context.Context, model *APIKeyResourceModel) (*client.CreateAPIKeyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.CreateAPIKeyRequest{
		Name:      model.Name.ValueString(),
		ExpiresAt: model.ExpiresAt.ValueStringPointer(),
	}
	caps, d := buildCapabilities(ctx, model.Capabilities)
	diags.Append(d...)
	req.Capabilities = caps
	roleIDs, d := buildRoleIDs(ctx, model.RoleIDs)
	diags.Append(d...)
	req.RoleIDs = roleIDs
	return req, diags
}

// buildUpdateRequest constructs an UpdateAPIKeyRequest from the planned changes.
func buildUpdateRequest(ctx context.Context, plan, state *APIKeyResourceModel) (*client.UpdateAPIKeyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateAPIKeyRequest{}

	if !plan.Name.Equal(state.Name) {
		req.Name = plan.Name.ValueStringPointer()
	}
	if !plan.Capabilities.Equal(state.Capabilities) {
		caps, d := buildCapabilities(ctx, plan.Capabilities)
		diags.Append(d...)
		req.Capabilities = caps
	}
	if !plan.RoleIDs.Equal(state.RoleIDs) {
		roleIDs, d := buildRoleIDs(ctx, plan.RoleIDs)
		diags.Append(d...)
		req.RoleIDs = roleIDs
	}

	return req, diags
}

// keepOrder returns prior when it holds the same elements as got, so a
// server-side reordering does not show up as a diff.
func keepOrder[T comparable](prior, got []T) []T {
	if len(prior) != len(got) {
		return got
	}
	counts := make(map[T]int, len(got))
	for _, v := range got {
		counts[v]++
	}
	for _, v := range prior {
		if counts[v] == 0 {
			return got
		}
		counts[v]--
	}
	return prior
}

func sameInstant(a, b string) bool {
	ta, errA := time.Parse(time.RFC3339Nano, a)
	tb, errB := time.Parse(time.RFC3339Nano, b)
	return errA == nil && errB == nil && ta.Equal(tb)
}

// optionalString converts a *string to a types.String, returning null for nil pointers.
func optionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
