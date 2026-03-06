package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationEnvVarResource{}
	_ resource.ResourceWithConfigure   = &applicationEnvVarResource{}
	_ resource.ResourceWithImportState = &applicationEnvVarResource{}
)

type applicationEnvVarResource struct {
	client *client.SevallaClient
}

type ApplicationEnvVarResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
	IsRuntime     types.Bool   `tfsdk:"is_runtime"`
	IsBuildtime   types.Bool   `tfsdk:"is_buildtime"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewEnvironmentVariableResource() resource.Resource {
	return &applicationEnvVarResource{}
}

func (r *applicationEnvVarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_environment_variable"
}

func (r *applicationEnvVarResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an environment variable for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment variable.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this environment variable belongs to.",
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
			"is_runtime": schema.BoolAttribute{
				Description: "Whether the environment variable is available at runtime.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_buildtime": schema.BoolAttribute{
				Description: "Whether the environment variable is available at build time.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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

func (r *applicationEnvVarResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationEnvVarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationEnvVarResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	createReq := &client.CreateEnvironmentVariableRequest{
		Key:   plan.Key.ValueString(),
		Value: plan.Value.ValueString(),
	}
	if !plan.IsRuntime.IsNull() && !plan.IsRuntime.IsUnknown() {
		v := plan.IsRuntime.ValueBool()
		createReq.IsRuntime = &v
	}
	if !plan.IsBuildtime.IsNull() && !plan.IsBuildtime.IsUnknown() {
		v := plan.IsBuildtime.ValueBool()
		createReq.IsBuildtime = &v
	}
	envVar, err := r.client.CreateEnvironmentVariable(ctx, "/applications", appID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Environment Variable", err.Error())
		return
	}

	flattenEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationEnvVarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	envVars, err := r.client.ListEnvironmentVariables(ctx, "/applications", appID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application Environment Variable", err.Error())
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

	flattenEnvVar(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationEnvVarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationEnvVarResourceModel
	var state ApplicationEnvVarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	updateReq := &client.UpdateEnvironmentVariableRequest{}

	if !plan.Key.Equal(state.Key) {
		v := plan.Key.ValueString()
		updateReq.Key = &v
	}
	if !plan.Value.Equal(state.Value) {
		v := plan.Value.ValueString()
		updateReq.Value = &v
	}
	if !plan.IsRuntime.IsNull() && !plan.IsRuntime.IsUnknown() && !plan.IsRuntime.Equal(state.IsRuntime) {
		v := plan.IsRuntime.ValueBool()
		updateReq.IsRuntime = &v
	}
	if !plan.IsBuildtime.IsNull() && !plan.IsBuildtime.IsUnknown() && !plan.IsBuildtime.Equal(state.IsBuildtime) {
		v := plan.IsBuildtime.ValueBool()
		updateReq.IsBuildtime = &v
	}

	envVar, err := r.client.UpdateEnvironmentVariable(ctx, "/applications", appID, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application Environment Variable", err.Error())
		return
	}

	flattenEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationEnvVarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironmentVariable(ctx, "/applications", state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Environment Variable", err.Error())
		return
	}
}

func (r *applicationEnvVarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'application_id/env_var_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenEnvVar(envVar *client.EnvironmentVariable, model *ApplicationEnvVarResourceModel) {
	model.ID = types.StringValue(envVar.ID)
	model.Key = types.StringValue(envVar.Key)
	model.Value = types.StringValue(envVar.Value)

	if envVar.IsRuntime != nil {
		model.IsRuntime = types.BoolValue(*envVar.IsRuntime)
	} else {
		model.IsRuntime = types.BoolNull()
	}
	if envVar.IsBuildtime != nil {
		model.IsBuildtime = types.BoolValue(*envVar.IsBuildtime)
	} else {
		model.IsBuildtime = types.BoolNull()
	}
	model.CreatedAt = optionalString(envVar.CreatedAt)
	model.UpdatedAt = optionalString(envVar.UpdatedAt)
}
