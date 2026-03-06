package globalenvvar

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &globalEnvVarResource{}
	_ resource.ResourceWithConfigure   = &globalEnvVarResource{}
	_ resource.ResourceWithImportState = &globalEnvVarResource{}
)

type globalEnvVarResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &globalEnvVarResource{}
}

func (r *globalEnvVarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_environment_variable"
}

func (r *globalEnvVarResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla global environment variable.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the global environment variable.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The key of the environment variable.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the environment variable.",
				Required:    true,
				Sensitive:   true,
			},
			"is_runtime": schema.BoolAttribute{
				Description: "Whether this environment variable is available at runtime.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_buildtime": schema.BoolAttribute{
				Description: "Whether this environment variable is available at build time.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the global environment variable was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the global environment variable was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *globalEnvVarResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.SevallaClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.SevallaClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *globalEnvVarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GlobalEnvVarResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildCreateRequest(&plan)

	envVar, err := r.client.CreateGlobalEnvVar(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Global Environment Variable",
			"Could not create global environment variable, unexpected error: "+err.Error(),
		)
		return
	}

	flattenGlobalEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalEnvVarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API does not have a get-by-ID endpoint; list all and find by ID.
	envVars, err := r.client.ListGlobalEnvVars(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Global Environment Variables",
			"Could not list global environment variables: "+err.Error(),
		)
		return
	}

	var found *client.GlobalEnvironmentVariable
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

	flattenGlobalEnvVar(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *globalEnvVarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalEnvVarResourceModel
	var state GlobalEnvVarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildUpdateRequest(&plan)

	envVar, err := r.client.UpdateGlobalEnvVar(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Global Environment Variable",
			"Could not update global environment variable ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	flattenGlobalEnvVar(envVar, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalEnvVarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalEnvVarResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlobalEnvVar(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Global Environment Variable",
			"Could not delete global environment variable ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *globalEnvVarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
