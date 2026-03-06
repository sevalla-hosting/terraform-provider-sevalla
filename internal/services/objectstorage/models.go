package objectstorage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// ObjectStorageResourceModel is the Terraform state model for the sevalla_object_storage resource.
type ObjectStorageResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DisplayName  types.String `tfsdk:"display_name"`
	Location     types.String `tfsdk:"location"`
	Jurisdiction types.String `tfsdk:"jurisdiction"`
	PublicAccess types.Bool   `tfsdk:"public_access"`
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Domain       types.String `tfsdk:"domain"`
	Endpoint     types.String `tfsdk:"endpoint"`
	AccessKey    types.String `tfsdk:"access_key"`
	SecretKey    types.String `tfsdk:"secret_key"`
	BucketName   types.String `tfsdk:"bucket_name"`
	CompanyID    types.String `tfsdk:"company_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// flattenObjectStorage maps a client.ObjectStorage to the Terraform resource model.
func flattenObjectStorage(_ context.Context, os *client.ObjectStorage, model *ObjectStorageResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(os.ID)
	model.Name = types.StringValue(os.Name)
	model.DisplayName = types.StringValue(os.DisplayName)
	model.Location = types.StringValue(os.Location)
	model.Jurisdiction = types.StringValue(os.Jurisdiction)
	model.BucketName = types.StringValue(os.BucketName)
	model.CompanyID = types.StringValue(os.CompanyID)
	model.CreatedAt = types.StringValue(os.CreatedAt)
	model.UpdatedAt = types.StringValue(os.UpdatedAt)

	model.ProjectID = optionalString(os.ProjectID)
	model.Domain = optionalString(os.Domain)
	model.Endpoint = optionalString(os.Endpoint)
	model.AccessKey = optionalString(os.AccessKey)
	model.SecretKey = optionalString(os.SecretKey)

	return diags
}

// buildCreateRequest constructs a CreateObjectStorageRequest from the Terraform plan model.
func buildCreateRequest(model *ObjectStorageResourceModel) *client.CreateObjectStorageRequest {
	req := &client.CreateObjectStorageRequest{
		DisplayName: model.DisplayName.ValueString(),
	}

	if !model.Location.IsNull() && !model.Location.IsUnknown() {
		v := model.Location.ValueString()
		req.Location = &v
	}
	if !model.Jurisdiction.IsNull() && !model.Jurisdiction.IsUnknown() {
		v := model.Jurisdiction.ValueString()
		req.Jurisdiction = &v
	}
	if !model.PublicAccess.IsNull() && !model.PublicAccess.IsUnknown() {
		v := model.PublicAccess.ValueBool()
		req.PublicAccess = &v
	}
	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateObjectStorageRequest from the Terraform plan model.
func buildUpdateRequest(plan *ObjectStorageResourceModel, state *ObjectStorageResourceModel) *client.UpdateObjectStorageRequest {
	req := &client.UpdateObjectStorageRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.ProjectID.Equal(state.ProjectID) {
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
