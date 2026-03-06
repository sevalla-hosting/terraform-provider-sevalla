package pipeline

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// PipelineResourceModel is the Terraform state model for the sevalla_pipeline resource.
type PipelineResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Type        types.String `tfsdk:"type"`
	ProjectID   types.String `tfsdk:"project_id"`
	CompanyID   types.String `tfsdk:"company_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// flattenPipeline maps a client.Pipeline to the Terraform resource model.
func flattenPipeline(p *client.Pipeline, model *PipelineResourceModel) {
	model.ID = types.StringValue(p.ID)
	model.DisplayName = types.StringValue(p.DisplayName)
	model.Type = types.StringValue(p.Type)
	model.ProjectID = optionalString(p.ProjectID)
	model.CompanyID = optionalString(p.CompanyID)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)
}

// buildCreateRequest constructs a CreatePipelineRequest from the Terraform plan model.
func buildCreateRequest(model *PipelineResourceModel) *client.CreatePipelineRequest {
	req := &client.CreatePipelineRequest{
		DisplayName: model.DisplayName.ValueString(),
		Type:        model.Type.ValueString(),
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdatePipelineRequest from the Terraform plan model.
func buildUpdateRequest(plan *PipelineResourceModel, state *PipelineResourceModel) *client.UpdatePipelineRequest {
	req := &client.UpdatePipelineRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}

	if !plan.Type.Equal(state.Type) {
		v := plan.Type.ValueString()
		req.Type = &v
	}

	if !plan.ProjectID.Equal(state.ProjectID) {
		if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() {
			v := plan.ProjectID.ValueString()
			req.ProjectID = &v
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
