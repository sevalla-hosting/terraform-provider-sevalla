package staticsite

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

// StaticSiteResourceModel is the Terraform state model for the sevalla_static_site resource.
type StaticSiteResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	DisplayName        types.String `tfsdk:"display_name"`
	Source             types.String `tfsdk:"source"`
	GitType            types.String `tfsdk:"git_type"`
	RepoURL            types.String `tfsdk:"repo_url"`
	DefaultBranch      types.String `tfsdk:"default_branch"`
	ProjectID          types.String `tfsdk:"project_id"`
	AutoDeploy         types.Bool   `tfsdk:"auto_deploy"`
	IsPreviewEnabled   types.Bool   `tfsdk:"is_preview_enabled"`
	InstallCommand     types.String `tfsdk:"install_command"`
	BuildCommand       types.String `tfsdk:"build_command"`
	PublishedDirectory types.String `tfsdk:"published_directory"`
	RootDirectory      types.String `tfsdk:"root_directory"`
	NodeVersion        types.String `tfsdk:"node_version"`
	IndexFile          types.String `tfsdk:"index_file"`
	ErrorFile          types.String `tfsdk:"error_file"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	Hostname           types.String `tfsdk:"hostname"`
	CompanyID          types.String `tfsdk:"company_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

// StaticSiteListItemModel is the Terraform state model for items in the sevalla_static_sites data source.
type StaticSiteListItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Status      types.String `tfsdk:"status"`
	Source      types.String `tfsdk:"source"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// StaticSiteListDataSourceModel is the top-level model for the sevalla_static_sites data source.
type StaticSiteListDataSourceModel struct {
	StaticSites []StaticSiteListItemModel `tfsdk:"static_sites"`
}

// flattenStaticSite maps a client.StaticSite to the Terraform resource model.
func flattenStaticSite(_ context.Context, ss *client.StaticSite, model *StaticSiteResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(ss.ID)
	model.Name = types.StringValue(ss.Name)
	model.DisplayName = types.StringValue(ss.DisplayName)
	model.Source = types.StringValue(ss.Source)
	model.AutoDeploy = types.BoolValue(ss.AutoDeploy)
	model.IsPreviewEnabled = types.BoolValue(ss.IsPreviewEnabled)
	model.CreatedAt = types.StringValue(ss.CreatedAt)
	model.UpdatedAt = types.StringValue(ss.UpdatedAt)

	// Nullable string fields
	model.CompanyID = optionalString(ss.CompanyID)
	model.ProjectID = optionalString(ss.ProjectID)
	model.Status = optionalString(ss.Status)
	model.GitType = optionalString(ss.GitType)
	model.RepoURL = optionalString(ss.RepoURL)
	model.DefaultBranch = optionalString(ss.DefaultBranch)
	model.Hostname = optionalString(ss.Hostname)
	model.InstallCommand = optionalString(ss.InstallCommand)
	model.BuildCommand = optionalString(ss.BuildCommand)
	model.PublishedDirectory = optionalString(ss.PublishedDirectory)
	model.RootDirectory = optionalString(ss.RootDirectory)
	model.NodeVersion = optionalString(ss.NodeVersion)
	model.IndexFile = optionalString(ss.IndexFile)
	model.ErrorFile = optionalString(ss.ErrorFile)

	return diags
}

// buildCreateRequest constructs a CreateStaticSiteRequest from the Terraform plan model.
func buildCreateRequest(model *StaticSiteResourceModel) *client.CreateStaticSiteRequest {
	req := &client.CreateStaticSiteRequest{
		DisplayName:   model.DisplayName.ValueString(),
		RepoURL:       model.RepoURL.ValueString(),
		DefaultBranch: model.DefaultBranch.ValueString(),
	}

	if !model.Source.IsNull() && !model.Source.IsUnknown() {
		v := model.Source.ValueString()
		req.Source = &v
	}
	if !model.GitType.IsNull() && !model.GitType.IsUnknown() {
		v := model.GitType.ValueString()
		req.GitType = &v
	}
	if !model.AutoDeploy.IsNull() && !model.AutoDeploy.IsUnknown() {
		v := model.AutoDeploy.ValueBool()
		req.AutoDeploy = &v
	}
	if !model.IsPreviewEnabled.IsNull() && !model.IsPreviewEnabled.IsUnknown() {
		v := model.IsPreviewEnabled.ValueBool()
		req.IsPreviewEnabled = &v
	}
	if !model.InstallCommand.IsNull() && !model.InstallCommand.IsUnknown() {
		v := model.InstallCommand.ValueString()
		req.InstallCommand = &v
	}
	if !model.BuildCommand.IsNull() && !model.BuildCommand.IsUnknown() {
		v := model.BuildCommand.ValueString()
		req.BuildCommand = &v
	}
	if !model.PublishedDirectory.IsNull() && !model.PublishedDirectory.IsUnknown() {
		v := model.PublishedDirectory.ValueString()
		req.PublishedDirectory = &v
	}
	if !model.RootDirectory.IsNull() && !model.RootDirectory.IsUnknown() {
		v := model.RootDirectory.ValueString()
		req.RootDirectory = &v
	}
	if !model.NodeVersion.IsNull() && !model.NodeVersion.IsUnknown() {
		v := model.NodeVersion.ValueString()
		req.NodeVersion = &v
	}
	if !model.IndexFile.IsNull() && !model.IndexFile.IsUnknown() {
		v := model.IndexFile.ValueString()
		req.IndexFile = &v
	}
	if !model.ErrorFile.IsNull() && !model.ErrorFile.IsUnknown() {
		v := model.ErrorFile.ValueString()
		req.ErrorFile = &v
	}
	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		v := model.ProjectID.ValueString()
		req.ProjectID = &v
	}

	return req
}

// buildUpdateRequest constructs an UpdateStaticSiteRequest from the Terraform plan model.
func buildUpdateRequest(_ context.Context, plan *StaticSiteResourceModel, state *StaticSiteResourceModel) *client.UpdateStaticSiteRequest {
	req := &client.UpdateStaticSiteRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.AutoDeploy.IsNull() && !plan.AutoDeploy.IsUnknown() && !plan.AutoDeploy.Equal(state.AutoDeploy) {
		v := plan.AutoDeploy.ValueBool()
		req.AutoDeploy = &v
	}
	if !plan.DefaultBranch.IsNull() && !plan.DefaultBranch.IsUnknown() && !plan.DefaultBranch.Equal(state.DefaultBranch) {
		v := plan.DefaultBranch.ValueString()
		req.DefaultBranch = &v
	}
	if !plan.BuildCommand.IsNull() && !plan.BuildCommand.IsUnknown() && !plan.BuildCommand.Equal(state.BuildCommand) {
		v := plan.BuildCommand.ValueString()
		req.BuildCommand = &v
	}
	if !plan.NodeVersion.IsNull() && !plan.NodeVersion.IsUnknown() && !plan.NodeVersion.Equal(state.NodeVersion) {
		v := plan.NodeVersion.ValueString()
		req.NodeVersion = &v
	}
	if !plan.PublishedDirectory.IsNull() && !plan.PublishedDirectory.IsUnknown() && !plan.PublishedDirectory.Equal(state.PublishedDirectory) {
		v := plan.PublishedDirectory.ValueString()
		req.PublishedDirectory = &v
	}
	if !plan.IsPreviewEnabled.IsNull() && !plan.IsPreviewEnabled.IsUnknown() && !plan.IsPreviewEnabled.Equal(state.IsPreviewEnabled) {
		v := plan.IsPreviewEnabled.ValueBool()
		req.IsPreviewEnabled = &v
	}
	if !plan.Source.IsNull() && !plan.Source.IsUnknown() && !plan.Source.Equal(state.Source) {
		v := plan.Source.ValueString()
		req.Source = &v
	}
	if !plan.GitType.IsNull() && !plan.GitType.IsUnknown() && !plan.GitType.Equal(state.GitType) {
		v := plan.GitType.ValueString()
		req.GitType = &v
	}
	if !plan.RepoURL.IsNull() && !plan.RepoURL.IsUnknown() && !plan.RepoURL.Equal(state.RepoURL) {
		v := plan.RepoURL.ValueString()
		req.RepoURL = &v
	}
	if !plan.InstallCommand.IsNull() && !plan.InstallCommand.IsUnknown() && !plan.InstallCommand.Equal(state.InstallCommand) {
		v := plan.InstallCommand.ValueString()
		req.InstallCommand = &v
	}
	if !plan.RootDirectory.IsNull() && !plan.RootDirectory.IsUnknown() && !plan.RootDirectory.Equal(state.RootDirectory) {
		v := plan.RootDirectory.ValueString()
		req.RootDirectory = &v
	}
	if !plan.IndexFile.IsNull() && !plan.IndexFile.IsUnknown() && !plan.IndexFile.Equal(state.IndexFile) {
		v := plan.IndexFile.ValueString()
		req.IndexFile = &v
	}
	if !plan.ErrorFile.IsNull() && !plan.ErrorFile.IsUnknown() && !plan.ErrorFile.Equal(state.ErrorFile) {
		v := plan.ErrorFile.ValueString()
		req.ErrorFile = &v
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
