package project

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// ProjectResourceModel is the Terraform state model for the sevalla_project resource.
type ProjectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Name        types.String `tfsdk:"name"`
	CompanyID   types.String `tfsdk:"company_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// flattenProject maps a client.Project to the Terraform resource model.
func flattenProject(p *client.Project, model *ProjectResourceModel) {
	model.ID = types.StringValue(p.ID)
	model.DisplayName = types.StringValue(p.DisplayName)
	model.Name = types.StringValue(p.Name)
	model.CompanyID = optionalString(p.CompanyID)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)
}

// buildCreateRequest constructs a CreateProjectRequest from the Terraform plan model.
func buildCreateRequest(model *ProjectResourceModel) *client.CreateProjectRequest {
	return &client.CreateProjectRequest{
		DisplayName: model.DisplayName.ValueString(),
	}
}

// buildUpdateRequest constructs an UpdateProjectRequest from the Terraform plan model.
func buildUpdateRequest(plan *ProjectResourceModel, state *ProjectResourceModel) *client.UpdateProjectRequest {
	req := &client.UpdateProjectRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
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
