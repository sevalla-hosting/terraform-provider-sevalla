package dockerregistry

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// DockerRegistryResourceModel is the Terraform state model for the sevalla_docker_registry resource.
type DockerRegistryResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Registry  types.String `tfsdk:"registry"`
	Username  types.String `tfsdk:"username"`
	Secret    types.String `tfsdk:"secret"`
	CompanyID types.String `tfsdk:"company_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// flattenDockerRegistry maps a client.DockerRegistry to the Terraform resource model.
// The secret field is write-only and not returned by the API, so it is preserved from state.
func flattenDockerRegistry(r *client.DockerRegistry, model *DockerRegistryResourceModel) {
	model.ID = types.StringValue(r.ID)
	model.Name = types.StringValue(r.Name)
	model.Registry = optionalString(r.Registry)
	model.Username = optionalString(r.Username)
	// Secret is write-only; not returned by the API. Preserve existing state value.
	model.CompanyID = optionalString(r.CompanyID)
	model.CreatedAt = types.StringValue(r.CreatedAt)
	model.UpdatedAt = types.StringValue(r.UpdatedAt)
}

// buildCreateRequest constructs a CreateDockerRegistryRequest from the Terraform plan model.
func buildCreateRequest(model *DockerRegistryResourceModel) *client.CreateDockerRegistryRequest {
	req := &client.CreateDockerRegistryRequest{
		Name:     model.Name.ValueString(),
		Username: model.Username.ValueString(),
		Secret:   model.Secret.ValueString(),
	}

	if !model.Registry.IsNull() && !model.Registry.IsUnknown() {
		v := model.Registry.ValueString()
		req.Registry = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateDockerRegistryRequest from the Terraform plan model.
func buildUpdateRequest(plan *DockerRegistryResourceModel, state *DockerRegistryResourceModel) *client.UpdateDockerRegistryRequest {
	req := &client.UpdateDockerRegistryRequest{}

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		req.Name = &v
	}
	if !plan.Registry.IsNull() && !plan.Registry.IsUnknown() && !plan.Registry.Equal(state.Registry) {
		v := plan.Registry.ValueString()
		req.Registry = &v
	}
	if !plan.Username.IsNull() && !plan.Username.IsUnknown() && !plan.Username.Equal(state.Username) {
		v := plan.Username.ValueString()
		req.Username = &v
	}
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() && !plan.Secret.Equal(state.Secret) {
		v := plan.Secret.ValueString()
		req.Secret = &v
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
