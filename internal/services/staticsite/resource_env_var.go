package staticsite

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &staticSiteEnvVarResource{}
	_ resource.ResourceWithConfigure   = &staticSiteEnvVarResource{}
	_ resource.ResourceWithImportState = &staticSiteEnvVarResource{}
)

type staticSiteEnvVarResource struct {
	client *client.SevallaClient
}

type StaticSiteEnvVarResourceModel struct {
	ID           types.String `tfsdk:"id"`
	StaticSiteID types.String `tfsdk:"static_site_id"`
	Key          types.String `tfsdk:"key"`
	Value        types.String `tfsdk:"value"`
	IsProduction types.Bool   `tfsdk:"is_production"`
	IsPreview    types.Bool   `tfsdk:"is_preview"`
	Branch       types.String `tfsdk:"branch"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewEnvironmentVariableResource() resource.Resource {
	return &staticSiteEnvVarResource{}
}

func (r *staticSiteEnvVarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site_environment_variable"
}

func (r *staticSiteEnvVarResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an environment variable for a Sevalla static site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment variable.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"static_site_id": schema.StringAttribute{
				Description: "The ID of the static site this environment variable belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The environment variable key.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The environment variable value.",
				Required:    true,
				Sensitive:   true,
			},
			"is_production": schema.BoolAttribute{
				Description: "Whether this variable is used in production deployments. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			"is_preview": schema.BoolAttribute{
				Description: "Whether this variable is used in preview deployments. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			"branch": schema.StringAttribute{
				Description: "Git branch this variable is scoped to. When omitted, the variable applies to all branches.",
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the environment variable was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the environment variable was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *staticSiteEnvVarResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.SevallaClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.SevallaClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *staticSiteEnvVarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticSiteEnvVarResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := plan.StaticSiteID.ValueString()
	createReq := &client.CreateEnvironmentVariableRequest{
		Key:   plan.Key.ValueString(),
		Value: plan.Value.ValueString(),
	}
	if !plan.IsProduction.IsNull() && !plan.IsProduction.IsUnknown() {
		v := plan.IsProduction.ValueBool()
		createReq.IsProduction = &v
	}
	if !plan.IsPreview.IsNull() && !plan.IsPreview.IsUnknown() {
		v := plan.IsPreview.ValueBool()
		createReq.IsPreview = &v
	}
	if !plan.Branch.IsNull() && !plan.Branch.IsUnknown() {
		v := plan.Branch.ValueString()
		createReq.Branch = &v
	}

	envVar, err := r.client.CreateEnvironmentVariable(ctx, "/static-sites", siteID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Static Site Environment Variable", err.Error())
		return
	}

	flattenStaticSiteEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticSiteEnvVarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticSiteEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.StaticSiteID.ValueString()
	envVars, err := r.client.ListEnvironmentVariables(ctx, "/static-sites", siteID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Static Site Environment Variable", err.Error())
		return
	}

	var found *client.EnvironmentVariable
	for i := range envVars {
		if envVars[i].ID == state.ID.ValueString() {
			found = &envVars[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenStaticSiteEnvVar(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *staticSiteEnvVarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticSiteEnvVarResourceModel
	var state StaticSiteEnvVarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.StaticSiteID.ValueString()
	updateReq := &client.UpdateEnvironmentVariableRequest{}

	if !plan.Key.Equal(state.Key) {
		v := plan.Key.ValueString()
		updateReq.Key = &v
	}
	if !plan.Value.Equal(state.Value) {
		v := plan.Value.ValueString()
		updateReq.Value = &v
	}
	if !plan.IsProduction.Equal(state.IsProduction) {
		v := plan.IsProduction.ValueBool()
		updateReq.IsProduction = &v
	}
	if !plan.IsPreview.Equal(state.IsPreview) {
		v := plan.IsPreview.ValueBool()
		updateReq.IsPreview = &v
	}
	if !plan.Branch.Equal(state.Branch) {
		if plan.Branch.IsNull() {
			updateReq.Branch = nil
		} else {
			v := plan.Branch.ValueString()
			updateReq.Branch = &v
		}
	}

	envVar, err := r.client.UpdateEnvironmentVariable(ctx, "/static-sites", siteID, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Static Site Environment Variable", err.Error())
		return
	}

	flattenStaticSiteEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticSiteEnvVarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticSiteEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironmentVariable(ctx, "/static-sites", state.StaticSiteID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Static Site Environment Variable", err.Error())
		return
	}
}

func (r *staticSiteEnvVarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'static_site_id/env_var_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenStaticSiteEnvVar(envVar *client.EnvironmentVariable, model *StaticSiteEnvVarResourceModel) {
	model.ID = types.StringValue(envVar.ID)
	model.Key = types.StringValue(envVar.Key)
	model.Value = types.StringValue(envVar.Value)

	if envVar.IsProduction != nil {
		model.IsProduction = types.BoolValue(*envVar.IsProduction)
	} else {
		model.IsProduction = types.BoolNull()
	}
	if envVar.IsPreview != nil {
		model.IsPreview = types.BoolValue(*envVar.IsPreview)
	} else {
		model.IsPreview = types.BoolNull()
	}
	if envVar.Branch != nil {
		model.Branch = types.StringValue(*envVar.Branch)
	} else {
		model.Branch = types.StringNull()
	}
	if envVar.CreatedAt != nil {
		model.CreatedAt = types.StringValue(*envVar.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
	if envVar.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(*envVar.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}
}
