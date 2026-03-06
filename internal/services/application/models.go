package application

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// ApplicationResourceModel is the Terraform state model for the sevalla_application resource.
type ApplicationResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	DisplayName                types.String `tfsdk:"display_name"`
	ClusterID                  types.String `tfsdk:"cluster_id"`
	Source                     types.String `tfsdk:"source"`
	ProjectID                  types.String `tfsdk:"project_id"`
	GitType                    types.String `tfsdk:"git_type"`
	RepoURL                    types.String `tfsdk:"repo_url"`
	DefaultBranch              types.String `tfsdk:"default_branch"`
	DockerImage                types.String `tfsdk:"docker_image"`
	DockerRegistryCredentialID types.String `tfsdk:"docker_registry_credential_id"`
	AutoDeploy                 types.Bool   `tfsdk:"auto_deploy"`
	BuildType                  types.String `tfsdk:"build_type"`
	BuildPath                  types.String `tfsdk:"build_path"`
	BuildCacheEnabled          types.Bool   `tfsdk:"build_cache_enabled"`
	HibernationEnabled         types.Bool   `tfsdk:"hibernation_enabled"`
	HibernateAfterSeconds      types.Int64  `tfsdk:"hibernate_after_seconds"`
	DockerfilePath             types.String `tfsdk:"dockerfile_path"`
	DockerContext              types.String `tfsdk:"docker_context"`
	PackBuilder                types.String `tfsdk:"pack_builder"`
	NixpacksVersion            types.String `tfsdk:"nixpacks_version"`
	AllowDeployPaths           types.List   `tfsdk:"allow_deploy_paths"`
	IgnoreDeployPaths          types.List   `tfsdk:"ignore_deploy_paths"`
	Buildpacks                 types.List   `tfsdk:"buildpacks"`
	WaitForChecks              types.Bool   `tfsdk:"wait_for_checks"`
	Name                       types.String `tfsdk:"name"`
	Namespace                  types.String `tfsdk:"namespace"`
	CompanyID                  types.String `tfsdk:"company_id"`
	Type                       types.String `tfsdk:"type"`
	Status                     types.String `tfsdk:"status"`
	IsSuspended                types.Bool   `tfsdk:"is_suspended"`
	CreatedBy                  types.String `tfsdk:"created_by"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	UpdatedAt                  types.String `tfsdk:"updated_at"`
}

// ApplicationListItemModel is the Terraform state model for items in the sevalla_applications data source.
type ApplicationListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
	Source      types.String `tfsdk:"source"`
	IsSuspended types.Bool   `tfsdk:"is_suspended"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// ApplicationListDataSourceModel is the top-level model for the sevalla_applications data source.
type ApplicationListDataSourceModel struct {
	Applications []ApplicationListItemModel `tfsdk:"applications"`
}

// flattenApplication maps a client.Application to the Terraform resource model.
// The model parameter is used to preserve values that are not returned by the API
// (e.g., cluster_id, docker_registry_credential_id).
func flattenApplication(ctx context.Context, app *client.Application, model *ApplicationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(app.ID)
	model.Name = types.StringValue(app.Name)
	model.DisplayName = types.StringValue(app.DisplayName)
	model.Source = types.StringValue(app.Source)
	model.Type = types.StringValue(app.Type)
	model.AutoDeploy = types.BoolValue(app.AutoDeploy)
	model.BuildCacheEnabled = types.BoolValue(app.BuildCacheEnabled)
	model.HibernationEnabled = types.BoolValue(app.HibernationEnabled)
	model.WaitForChecks = types.BoolValue(app.WaitForChecks)
	model.IsSuspended = types.BoolValue(app.IsSuspended)
	model.BuildType = types.StringValue(app.BuildType)
	model.CreatedAt = types.StringValue(app.CreatedAt)
	model.UpdatedAt = types.StringValue(app.UpdatedAt)

	// Nullable string fields
	model.Namespace = optionalString(app.Namespace)
	model.CompanyID = optionalString(app.CompanyID)
	model.ProjectID = optionalString(app.ProjectID)
	model.Status = optionalString(app.Status)
	model.GitType = optionalString(app.GitType)
	model.RepoURL = optionalString(app.RepoURL)
	model.DefaultBranch = optionalString(app.DefaultBranch)
	model.DockerImage = optionalString(app.DockerImage)

	// Preserve cluster_id and docker_registry_credential_id from plan/state
	// when the API does not return them (write-only / not echoed on GET).
	if app.ClusterID != nil {
		model.ClusterID = types.StringValue(*app.ClusterID)
	}
	if app.DockerRegistryCredentialID != nil {
		model.DockerRegistryCredentialID = types.StringValue(*app.DockerRegistryCredentialID)
	}
	model.BuildPath = optionalString(app.BuildPath)
	model.PackBuilder = optionalString(app.PackBuilder)
	model.NixpacksVersion = optionalString(app.NixpacksVersion)
	model.DockerfilePath = optionalString(app.DockerfilePath)
	model.DockerContext = optionalString(app.DockerContext)
	model.CreatedBy = optionalString(app.CreatedBy)

	// Nullable int64 field
	if app.HibernateAfterSeconds != nil {
		model.HibernateAfterSeconds = types.Int64Value(*app.HibernateAfterSeconds)
	} else {
		model.HibernateAfterSeconds = types.Int64Null()
	}

	// List fields
	allowPaths, d := types.ListValueFrom(ctx, types.StringType, app.AllowDeployPaths)
	diags.Append(d...)
	model.AllowDeployPaths = allowPaths

	ignorePaths, d := types.ListValueFrom(ctx, types.StringType, app.IgnoreDeployPaths)
	diags.Append(d...)
	model.IgnoreDeployPaths = ignorePaths

	// Buildpacks
	buildpackAttrTypes := map[string]attr.Type{
		"order":  types.Int64Type,
		"source": types.StringType,
	}
	if len(app.Buildpacks) > 0 {
		bpObjects := make([]attr.Value, 0, len(app.Buildpacks))
		for _, bp := range app.Buildpacks {
			obj, d := types.ObjectValue(buildpackAttrTypes, map[string]attr.Value{
				"order":  types.Int64Value(int64(bp.Order)),
				"source": types.StringValue(bp.Source),
			})
			diags.Append(d...)
			bpObjects = append(bpObjects, obj)
		}
		bpList, d := types.ListValue(types.ObjectType{AttrTypes: buildpackAttrTypes}, bpObjects)
		diags.Append(d...)
		model.Buildpacks = bpList
	} else {
		model.Buildpacks = types.ListNull(types.ObjectType{AttrTypes: buildpackAttrTypes})
	}

	return diags
}

// buildCreateRequest constructs a CreateApplicationRequest from the Terraform plan model.
func buildCreateRequest(model *ApplicationResourceModel) *client.CreateApplicationRequest {
	req := &client.CreateApplicationRequest{
		DisplayName: model.DisplayName.ValueString(),
		ClusterID:   model.ClusterID.ValueString(),
		Source:      model.Source.ValueString(),
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}
	if !model.GitType.IsNull() && !model.GitType.IsUnknown() {
		v := model.GitType.ValueString()
		req.GitType = &v
	}
	if !model.RepoURL.IsNull() && !model.RepoURL.IsUnknown() {
		v := model.RepoURL.ValueString()
		req.RepoURL = &v
	}
	if !model.DefaultBranch.IsNull() && !model.DefaultBranch.IsUnknown() {
		v := model.DefaultBranch.ValueString()
		req.DefaultBranch = &v
	}
	if !model.DockerImage.IsNull() && !model.DockerImage.IsUnknown() {
		v := model.DockerImage.ValueString()
		req.DockerImage = &v
	}
	if !model.DockerRegistryCredentialID.IsNull() && !model.DockerRegistryCredentialID.IsUnknown() {
		v := model.DockerRegistryCredentialID.ValueString()
		req.DockerRegistryCredentialID = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateApplicationRequest from the Terraform plan model.
func buildUpdateRequest(ctx context.Context, plan *ApplicationResourceModel, state *ApplicationResourceModel) *client.UpdateApplicationRequest {
	req := &client.UpdateApplicationRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.Source.Equal(state.Source) {
		v := plan.Source.ValueString()
		req.Source = &v
	}
	if !plan.AutoDeploy.IsNull() && !plan.AutoDeploy.IsUnknown() && !plan.AutoDeploy.Equal(state.AutoDeploy) {
		v := plan.AutoDeploy.ValueBool()
		req.AutoDeploy = &v
	}
	if !plan.DefaultBranch.IsNull() && !plan.DefaultBranch.IsUnknown() && !plan.DefaultBranch.Equal(state.DefaultBranch) {
		v := plan.DefaultBranch.ValueString()
		req.DefaultBranch = &v
	}
	if !plan.HibernationEnabled.IsNull() && !plan.HibernationEnabled.IsUnknown() && !plan.HibernationEnabled.Equal(state.HibernationEnabled) {
		v := plan.HibernationEnabled.ValueBool()
		req.HibernationEnabled = &v
	}
	if !plan.HibernateAfterSeconds.IsNull() && !plan.HibernateAfterSeconds.IsUnknown() && !plan.HibernateAfterSeconds.Equal(state.HibernateAfterSeconds) {
		v := plan.HibernateAfterSeconds.ValueInt64()
		req.HibernateAfterSeconds = &v
	}
	if !plan.BuildType.IsNull() && !plan.BuildType.IsUnknown() && !plan.BuildType.Equal(state.BuildType) {
		v := plan.BuildType.ValueString()
		req.BuildType = &v
	}
	if !plan.BuildPath.IsNull() && !plan.BuildPath.IsUnknown() && !plan.BuildPath.Equal(state.BuildPath) {
		v := plan.BuildPath.ValueString()
		req.BuildPath = &v
	}
	if !plan.DockerfilePath.IsNull() && !plan.DockerfilePath.IsUnknown() && !plan.DockerfilePath.Equal(state.DockerfilePath) {
		v := plan.DockerfilePath.ValueString()
		req.DockerfilePath = &v
	}
	if !plan.DockerContext.IsNull() && !plan.DockerContext.IsUnknown() && !plan.DockerContext.Equal(state.DockerContext) {
		v := plan.DockerContext.ValueString()
		req.DockerContext = &v
	}
	if !plan.BuildCacheEnabled.IsNull() && !plan.BuildCacheEnabled.IsUnknown() && !plan.BuildCacheEnabled.Equal(state.BuildCacheEnabled) {
		v := plan.BuildCacheEnabled.ValueBool()
		req.BuildCacheEnabled = &v
	}
	if !plan.DockerRegistryCredentialID.IsNull() && !plan.DockerRegistryCredentialID.IsUnknown() && !plan.DockerRegistryCredentialID.Equal(state.DockerRegistryCredentialID) {
		v := plan.DockerRegistryCredentialID.ValueString()
		req.DockerRegistryCredentialID = &v
	}
	if !plan.PackBuilder.IsNull() && !plan.PackBuilder.IsUnknown() && !plan.PackBuilder.Equal(state.PackBuilder) {
		v := plan.PackBuilder.ValueString()
		req.PackBuilder = &v
	}
	if !plan.NixpacksVersion.IsNull() && !plan.NixpacksVersion.IsUnknown() && !plan.NixpacksVersion.Equal(state.NixpacksVersion) {
		v := plan.NixpacksVersion.ValueString()
		req.NixpacksVersion = &v
	}
	if !plan.GitType.IsNull() && !plan.GitType.IsUnknown() && !plan.GitType.Equal(state.GitType) {
		v := plan.GitType.ValueString()
		req.GitType = &v
	}
	if !plan.RepoURL.IsNull() && !plan.RepoURL.IsUnknown() && !plan.RepoURL.Equal(state.RepoURL) {
		v := plan.RepoURL.ValueString()
		req.RepoURL = &v
	}
	if !plan.DockerImage.IsNull() && !plan.DockerImage.IsUnknown() && !plan.DockerImage.Equal(state.DockerImage) {
		v := plan.DockerImage.ValueString()
		req.DockerImage = &v
	}
	if !plan.AllowDeployPaths.Equal(state.AllowDeployPaths) {
		var paths []string
		plan.AllowDeployPaths.ElementsAs(ctx, &paths, false)
		req.AllowDeployPaths = paths
	}
	if !plan.IgnoreDeployPaths.Equal(state.IgnoreDeployPaths) {
		var paths []string
		plan.IgnoreDeployPaths.ElementsAs(ctx, &paths, false)
		req.IgnoreDeployPaths = paths
	}
	if !plan.Buildpacks.IsNull() && !plan.Buildpacks.IsUnknown() && !plan.Buildpacks.Equal(state.Buildpacks) {
		var bpObjects []types.Object
		plan.Buildpacks.ElementsAs(ctx, &bpObjects, false)
		buildpacks := make([]client.BuildpackConfig, 0, len(bpObjects))
		for _, obj := range bpObjects {
			attrs := obj.Attributes()
			buildpacks = append(buildpacks, client.BuildpackConfig{
				Order:  int(attrs["order"].(types.Int64).ValueInt64()),
				Source: attrs["source"].(types.String).ValueString(),
			})
		}
		req.Buildpacks = buildpacks
	}

	return req
}

// buildPostCreateUpdateRequest constructs an UpdateApplicationRequest for fields that the
// create endpoint does not accept. Returns nil if no update-only fields are set in the plan.
func buildPostCreateUpdateRequest(ctx context.Context, plan *ApplicationResourceModel) *client.UpdateApplicationRequest {
	req := &client.UpdateApplicationRequest{}
	hasFields := false

	if !plan.AutoDeploy.IsNull() && !plan.AutoDeploy.IsUnknown() {
		v := plan.AutoDeploy.ValueBool()
		req.AutoDeploy = &v
		hasFields = true
	}
	if !plan.BuildType.IsNull() && !plan.BuildType.IsUnknown() {
		v := plan.BuildType.ValueString()
		req.BuildType = &v
		hasFields = true
	}
	if !plan.BuildPath.IsNull() && !plan.BuildPath.IsUnknown() {
		v := plan.BuildPath.ValueString()
		req.BuildPath = &v
		hasFields = true
	}
	if !plan.BuildCacheEnabled.IsNull() && !plan.BuildCacheEnabled.IsUnknown() {
		v := plan.BuildCacheEnabled.ValueBool()
		req.BuildCacheEnabled = &v
		hasFields = true
	}
	if !plan.HibernationEnabled.IsNull() && !plan.HibernationEnabled.IsUnknown() {
		v := plan.HibernationEnabled.ValueBool()
		req.HibernationEnabled = &v
		hasFields = true
	}
	if !plan.HibernateAfterSeconds.IsNull() && !plan.HibernateAfterSeconds.IsUnknown() {
		v := plan.HibernateAfterSeconds.ValueInt64()
		req.HibernateAfterSeconds = &v
		hasFields = true
	}
	if !plan.DockerfilePath.IsNull() && !plan.DockerfilePath.IsUnknown() {
		v := plan.DockerfilePath.ValueString()
		req.DockerfilePath = &v
		hasFields = true
	}
	if !plan.DockerContext.IsNull() && !plan.DockerContext.IsUnknown() {
		v := plan.DockerContext.ValueString()
		req.DockerContext = &v
		hasFields = true
	}
	if !plan.PackBuilder.IsNull() && !plan.PackBuilder.IsUnknown() {
		v := plan.PackBuilder.ValueString()
		req.PackBuilder = &v
		hasFields = true
	}
	if !plan.NixpacksVersion.IsNull() && !plan.NixpacksVersion.IsUnknown() {
		v := plan.NixpacksVersion.ValueString()
		req.NixpacksVersion = &v
		hasFields = true
	}
	if !plan.AllowDeployPaths.IsNull() && !plan.AllowDeployPaths.IsUnknown() {
		var paths []string
		plan.AllowDeployPaths.ElementsAs(ctx, &paths, false)
		req.AllowDeployPaths = paths
		hasFields = true
	}
	if !plan.IgnoreDeployPaths.IsNull() && !plan.IgnoreDeployPaths.IsUnknown() {
		var paths []string
		plan.IgnoreDeployPaths.ElementsAs(ctx, &paths, false)
		req.IgnoreDeployPaths = paths
		hasFields = true
	}
	if !plan.Buildpacks.IsNull() && !plan.Buildpacks.IsUnknown() {
		var bpObjects []types.Object
		plan.Buildpacks.ElementsAs(ctx, &bpObjects, false)
		buildpacks := make([]client.BuildpackConfig, 0, len(bpObjects))
		for _, obj := range bpObjects {
			attrs := obj.Attributes()
			buildpacks = append(buildpacks, client.BuildpackConfig{
				Order:  int(attrs["order"].(types.Int64).ValueInt64()),
				Source: attrs["source"].(types.String).ValueString(),
			})
		}
		req.Buildpacks = buildpacks
		hasFields = true
	}

	if !hasFields {
		return nil
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
