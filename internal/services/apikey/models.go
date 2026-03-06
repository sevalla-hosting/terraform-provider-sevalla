package apikey

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// APIKeyResourceModel is the Terraform state model for the sevalla_api_key resource.
type APIKeyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
	Token        types.String `tfsdk:"token"`
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

// flattenAPIKey maps a client.APIKey to the Terraform resource model.
// The token field is only returned on create and is preserved from state on subsequent reads.
func flattenAPIKey(k *client.APIKey, model *APIKeyResourceModel) {
	model.ID = types.StringValue(k.ID)
	model.Name = types.StringValue(k.Name)
	model.ExpiresAt = optionalString(k.ExpiresAt)
	model.Enabled = types.BoolValue(k.Enabled)
	model.CompanyID = optionalString(k.CompanyID)
	model.Source = optionalString(k.Source)
	model.LastUsedAt = optionalString(k.LastUsedAt)
	model.CreatedAt = types.StringValue(k.CreatedAt)
	model.UpdatedAt = types.StringValue(k.UpdatedAt)

	// Token is only returned on create; preserve existing value on read/update.
	if k.Token != nil {
		model.Token = types.StringValue(*k.Token)
	}

	// role_ids is write-only (not returned by GET); preserve from plan/state.
	// If unset, default to null.
	if model.RoleIDs.IsUnknown() {
		model.RoleIDs = types.ListNull(types.StringType)
	}

	// Extract capabilities from response objects.
	if k.Capabilities != nil {
		caps := make([]APIKeyCapabilityModel, 0, len(k.Capabilities))
		for _, raw := range k.Capabilities {
			var cap struct {
				Permission string  `json:"permission"`
				IDResource *string `json:"id_resource"`
			}
			if err := json.Unmarshal(raw, &cap); err == nil && cap.Permission != "" {
				m := APIKeyCapabilityModel{
					Permission: types.StringValue(cap.Permission),
					IDResource: optionalString(cap.IDResource),
				}
				caps = append(caps, m)
			}
		}
		capList, _ := types.ListValueFrom(context.Background(), capabilityObjectType, caps)
		model.Capabilities = capList
	} else {
		model.Capabilities = types.ListNull(capabilityObjectType)
	}

	// Extract roles from response objects.
	if k.Roles != nil {
		roles := make([]APIKeyRoleModel, 0, len(k.Roles))
		for _, raw := range k.Roles {
			var role struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(raw, &role); err == nil && role.ID != "" {
				roles = append(roles, APIKeyRoleModel{
					ID:          types.StringValue(role.ID),
					Name:        types.StringValue(role.Name),
					Description: types.StringValue(role.Description),
				})
			}
		}
		roleList, _ := types.ListValueFrom(context.Background(), roleObjectType, roles)
		model.Roles = roleList
	} else {
		model.Roles = types.ListNull(roleObjectType)
	}
}

// buildCreateRequest constructs a CreateAPIKeyRequest from the Terraform plan model.
func buildCreateRequest(model *APIKeyResourceModel) *client.CreateAPIKeyRequest {
	req := &client.CreateAPIKeyRequest{
		Name: model.Name.ValueString(),
	}

	if !model.ExpiresAt.IsNull() && !model.ExpiresAt.IsUnknown() {
		v := model.ExpiresAt.ValueString()
		req.ExpiresAt = &v
	}

	if !model.Capabilities.IsNull() && !model.Capabilities.IsUnknown() {
		var caps []APIKeyCapabilityModel
		model.Capabilities.ElementsAs(context.Background(), &caps, false)
		for _, c := range caps {
			capReq := client.APIKeyCapabilityRequest{
				Permission: c.Permission.ValueString(),
			}
			if !c.IDResource.IsNull() && !c.IDResource.IsUnknown() {
				v := c.IDResource.ValueString()
				capReq.IDResource = &v
			}
			req.Capabilities = append(req.Capabilities, capReq)
		}
	}

	if !model.RoleIDs.IsNull() && !model.RoleIDs.IsUnknown() {
		var roleIDs []string
		model.RoleIDs.ElementsAs(context.Background(), &roleIDs, false)
		req.RoleIDs = roleIDs
	}

	return req
}

// buildUpdateRequest constructs an UpdateAPIKeyRequest from the Terraform plan model.
func buildUpdateRequest(plan *APIKeyResourceModel, state *APIKeyResourceModel) *client.UpdateAPIKeyRequest {
	req := &client.UpdateAPIKeyRequest{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		req.Name = &v
	}

	if !plan.Capabilities.Equal(state.Capabilities) {
		if !plan.Capabilities.IsNull() && !plan.Capabilities.IsUnknown() {
			var caps []APIKeyCapabilityModel
			plan.Capabilities.ElementsAs(context.Background(), &caps, false)
			for _, c := range caps {
				capReq := client.APIKeyCapabilityRequest{
					Permission: c.Permission.ValueString(),
				}
				if !c.IDResource.IsNull() && !c.IDResource.IsUnknown() {
					v := c.IDResource.ValueString()
					capReq.IDResource = &v
				}
				req.Capabilities = append(req.Capabilities, capReq)
			}
		}
	}

	if !plan.RoleIDs.Equal(state.RoleIDs) {
		if !plan.RoleIDs.IsNull() && !plan.RoleIDs.IsUnknown() {
			var roleIDs []string
			plan.RoleIDs.ElementsAs(context.Background(), &roleIDs, false)
			req.RoleIDs = roleIDs
		}
	}

	return req
}

// optionalString converts a *string to a types.String, returning null for nil pointers.
func optionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
