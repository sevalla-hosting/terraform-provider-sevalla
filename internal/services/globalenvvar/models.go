package globalenvvar

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// GlobalEnvVarResourceModel is the Terraform state model for the sevalla_global_environment_variable resource.
type GlobalEnvVarResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	IsRuntime   types.Bool   `tfsdk:"is_runtime"`
	IsBuildtime types.Bool   `tfsdk:"is_buildtime"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// flattenGlobalEnvVar maps a client.GlobalEnvironmentVariable to the Terraform resource model.
func flattenGlobalEnvVar(e *client.GlobalEnvironmentVariable, model *GlobalEnvVarResourceModel) {
	model.ID = types.StringValue(e.ID)
	model.Key = types.StringValue(e.Key)
	model.Value = types.StringValue(e.Value)
	model.IsRuntime = optionalBool(e.IsRuntime)
	model.IsBuildtime = optionalBool(e.IsBuildtime)
	model.CreatedAt = types.StringValue(e.CreatedAt)
	model.UpdatedAt = types.StringValue(e.UpdatedAt)
}

// buildCreateRequest constructs a CreateGlobalEnvVarRequest from the Terraform plan model.
func buildCreateRequest(model *GlobalEnvVarResourceModel) *client.CreateGlobalEnvVarRequest {
	req := &client.CreateGlobalEnvVarRequest{
		Key:   model.Key.ValueString(),
		Value: model.Value.ValueString(),
	}

	if !model.IsRuntime.IsNull() && !model.IsRuntime.IsUnknown() {
		v := model.IsRuntime.ValueBool()
		req.IsRuntime = &v
	}
	if !model.IsBuildtime.IsNull() && !model.IsBuildtime.IsUnknown() {
		v := model.IsBuildtime.ValueBool()
		req.IsBuildtime = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateGlobalEnvVarRequest from the Terraform plan model.
func buildUpdateRequest(plan *GlobalEnvVarResourceModel) *client.UpdateGlobalEnvVarRequest {
	req := &client.UpdateGlobalEnvVarRequest{
		Key:   optionalStringPtr(plan.Key),
		Value: optionalStringPtr(plan.Value),
	}

	if !plan.IsRuntime.IsNull() && !plan.IsRuntime.IsUnknown() {
		v := plan.IsRuntime.ValueBool()
		req.IsRuntime = &v
	}
	if !plan.IsBuildtime.IsNull() && !plan.IsBuildtime.IsUnknown() {
		v := plan.IsBuildtime.ValueBool()
		req.IsBuildtime = &v
	}

	return req
}

// optionalStringPtr returns a *string from a types.String, or nil if null/unknown.
func optionalStringPtr(s types.String) *string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	v := s.ValueString()
	return &v
}

// optionalBool converts a *bool to a types.Bool, returning null for nil pointers.
func optionalBool(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}
