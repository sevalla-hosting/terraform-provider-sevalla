package database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// DatabaseResourceModel is the Terraform state model for the sevalla_database resource.
type DatabaseResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	DisplayName        types.String `tfsdk:"display_name"`
	Type               types.String `tfsdk:"type"`
	Version            types.String `tfsdk:"version"`
	ClusterID          types.String `tfsdk:"cluster_id"`
	ResourceTypeID     types.String `tfsdk:"resource_type_id"`
	DbName             types.String `tfsdk:"db_name"`
	DbPassword         types.String `tfsdk:"db_password"`
	DbUser             types.String `tfsdk:"db_user"`
	ProjectID          types.String `tfsdk:"project_id"`
	Extensions         types.Object `tfsdk:"extensions"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	IsSuspended        types.Bool   `tfsdk:"is_suspended"`
	ClusterDisplayName types.String `tfsdk:"cluster_display_name"`
	ClusterLocation    types.String `tfsdk:"cluster_location"`
	ResourceTypeName   types.String `tfsdk:"resource_type_name"`
	CPULimit           types.Int64  `tfsdk:"cpu_limit"`
	MemoryLimit        types.Int64  `tfsdk:"memory_limit"`
	StorageSize        types.Int64  `tfsdk:"storage_size"`
	InternalHostname   types.String `tfsdk:"internal_hostname"`
	InternalPort       types.String `tfsdk:"internal_port"`
	ExternalHostname   types.String `tfsdk:"external_hostname"`
	ExternalPort       types.String `tfsdk:"external_port"`
	CompanyID          types.String `tfsdk:"company_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

// DatabaseListItemModel is the Terraform state model for items in the sevalla_databases data source.
type DatabaseListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Type        types.String `tfsdk:"type"`
	Status      types.String `tfsdk:"status"`
	IsSuspended types.Bool   `tfsdk:"is_suspended"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// DatabaseListDataSourceModel is the top-level model for the sevalla_databases data source.
type DatabaseListDataSourceModel struct {
	Databases []DatabaseListItemModel `tfsdk:"databases"`
}

// flattenDatabase maps a client.Database to the Terraform resource model.
// The model parameter is used to preserve values that are not returned by the API
// (e.g., cluster_id).
func flattenDatabase(_ context.Context, db *client.Database, model *DatabaseResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(db.ID)
	model.Name = types.StringValue(db.Name)
	model.DisplayName = types.StringValue(db.DisplayName)
	model.Type = types.StringValue(db.Type)
	model.ClusterID = types.StringValue(db.ClusterID)
	model.ResourceTypeID = optionalString(db.ResourceTypeID)
	model.IsSuspended = types.BoolValue(db.IsSuspended)
	model.CreatedAt = types.StringValue(db.CreatedAt)
	model.UpdatedAt = types.StringValue(db.UpdatedAt)

	// Nullable string fields
	model.CompanyID = optionalString(db.CompanyID)
	model.ProjectID = optionalString(db.ProjectID)
	model.Status = optionalString(db.Status)
	model.Version = optionalString(db.Version)
	model.ClusterDisplayName = optionalString(db.ClusterDisplayName)
	model.ClusterLocation = optionalString(db.ClusterLocation)
	model.ResourceTypeName = optionalString(db.ResourceTypeName)
	model.DbName = optionalString(db.DbName)
	model.InternalHostname = optionalString(db.InternalHostname)
	model.ExternalHostname = optionalString(db.ExternalHostname)

	// Nullable int fields
	model.CPULimit = optionalInt64(db.CPULimit)
	model.MemoryLimit = optionalInt64(db.MemoryLimit)
	model.StorageSize = optionalInt64(db.StorageSize)
	model.InternalPort = optionalString(db.InternalPort)
	model.ExternalPort = optionalString(db.ExternalPort)

	return diags
}

// buildCreateRequest constructs a CreateDatabaseRequest from the Terraform plan model.
func buildCreateRequest(model *DatabaseResourceModel) *client.CreateDatabaseRequest {
	req := &client.CreateDatabaseRequest{
		DisplayName:    model.DisplayName.ValueString(),
		Type:           model.Type.ValueString(),
		Version:        model.Version.ValueString(),
		ClusterID:      model.ClusterID.ValueString(),
		ResourceTypeID: model.ResourceTypeID.ValueString(),
		DbName:         model.DbName.ValueString(),
		DbPassword:     model.DbPassword.ValueString(),
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}

	if !model.DbUser.IsNull() && !model.DbUser.IsUnknown() {
		v := model.DbUser.ValueString()
		req.DbUser = &v
	}

	if !model.Extensions.IsNull() && !model.Extensions.IsUnknown() {
		attrs := model.Extensions.Attributes()
		extensions := &client.DatabaseExtensions{}
		if v, ok := attrs["enable_vector"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			b := v.ValueBool()
			extensions.EnableVector = &b
		}
		if v, ok := attrs["enable_postgis"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			b := v.ValueBool()
			extensions.EnablePostgis = &b
		}
		if v, ok := attrs["enable_cron"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			b := v.ValueBool()
			extensions.EnableCron = &b
		}
		req.Extensions = extensions
	}

	return req
}

// buildUpdateRequest constructs an UpdateDatabaseRequest from the Terraform plan model.
func buildUpdateRequest(plan *DatabaseResourceModel, state *DatabaseResourceModel) *client.UpdateDatabaseRequest {
	req := &client.UpdateDatabaseRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.ResourceTypeID.Equal(state.ResourceTypeID) {
		v := plan.ResourceTypeID.ValueString()
		req.ResourceTypeID = &v
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

// optionalInt64 converts a *int64 to a types.Int64, returning null for nil pointers.
func optionalInt64(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}
