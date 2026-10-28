package pipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	Stages      types.List   `tfsdk:"stages"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// PipelineStageModel is an element of the computed stages list.
type PipelineStageModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Type        types.String `tfsdk:"type"`
	Order       types.Int64  `tfsdk:"order"`
}

var stageObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"id":           types.StringType,
	"display_name": types.StringType,
	"type":         types.StringType,
	"order":        types.Int64Type,
}}

// flattenPipeline maps a client.Pipeline to the Terraform resource model.
func flattenPipeline(ctx context.Context, p *client.Pipeline, model *PipelineResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(p.ID)
	model.DisplayName = types.StringValue(p.DisplayName)
	model.Type = types.StringValue(p.Type)
	model.ProjectID = optionalString(p.ProjectID)
	model.CompanyID = optionalString(p.CompanyID)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)

	stages := make([]PipelineStageModel, len(p.Stages))
	for i, st := range p.Stages {
		stages[i] = PipelineStageModel{
			ID:          types.StringValue(st.ID),
			DisplayName: types.StringValue(st.DisplayName),
			Type:        types.StringValue(st.Type),
			Order:       types.Int64Value(int64(st.Order)),
		}
	}
	list, diags := types.ListValueFrom(ctx, stageObjectType, stages)
	model.Stages = list
	return diags
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

	if !plan.ProjectID.IsUnknown() && !plan.ProjectID.Equal(state.ProjectID) {
		req.ProjectID = client.NullableOf(plan.ProjectID.ValueStringPointer())
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
