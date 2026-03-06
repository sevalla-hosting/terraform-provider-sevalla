package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// LoadBalancerResourceModel is the Terraform state model for the sevalla_load_balancer resource.
type LoadBalancerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Type        types.String `tfsdk:"type"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	CompanyID   types.String `tfsdk:"company_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// LoadBalancerListItemModel is the Terraform state model for items in the sevalla_load_balancers data source.
type LoadBalancerListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Type        types.String `tfsdk:"type"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// LoadBalancerListDataSourceModel is the top-level model for the sevalla_load_balancers data source.
type LoadBalancerListDataSourceModel struct {
	LoadBalancers []LoadBalancerListItemModel `tfsdk:"load_balancers"`
}

// flattenLoadBalancer maps a client.LoadBalancer to the Terraform resource model.
func flattenLoadBalancer(_ context.Context, lb *client.LoadBalancer, model *LoadBalancerResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(lb.ID)
	model.Name = types.StringValue(lb.Name)
	model.DisplayName = types.StringValue(lb.DisplayName)
	model.CreatedAt = types.StringValue(lb.CreatedAt)
	model.UpdatedAt = types.StringValue(lb.UpdatedAt)

	// Nullable string fields
	model.CompanyID = optionalString(lb.CompanyID)
	model.ProjectID = optionalString(lb.ProjectID)
	model.Type = optionalString(lb.Type)

	return diags
}

// buildCreateRequest constructs a CreateLoadBalancerRequest from the Terraform plan model.
func buildCreateRequest(model *LoadBalancerResourceModel) *client.CreateLoadBalancerRequest {
	req := &client.CreateLoadBalancerRequest{
		DisplayName: model.DisplayName.ValueString(),
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}

	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		v := model.Type.ValueString()
		req.Type = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateLoadBalancerRequest from the Terraform plan model.
func buildUpdateRequest(plan *LoadBalancerResourceModel, state *LoadBalancerResourceModel) *client.UpdateLoadBalancerRequest {
	req := &client.UpdateLoadBalancerRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.Type.IsNull() && !plan.Type.IsUnknown() && !plan.Type.Equal(state.Type) {
		v := plan.Type.ValueString()
		req.Type = &v
	}
	if !plan.ProjectID.IsNull() && !plan.ProjectID.IsUnknown() && !plan.ProjectID.Equal(state.ProjectID) {
		v := plan.ProjectID.ValueString()
		req.ProjectID = &v
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
