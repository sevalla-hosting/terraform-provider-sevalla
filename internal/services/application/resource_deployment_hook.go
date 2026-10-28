package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationDeploymentHookResource{}
	_ resource.ResourceWithConfigure   = &applicationDeploymentHookResource{}
	_ resource.ResourceWithImportState = &applicationDeploymentHookResource{}
)

type applicationDeploymentHookResource struct {
	client *client.SevallaClient
}

type ApplicationDeploymentHookResourceModel struct {
	ApplicationID types.String `tfsdk:"application_id"`
	URL           types.String `tfsdk:"url"`
}

func NewDeploymentHookResource() resource.Resource {
	return &applicationDeploymentHookResource{}
}

func (r *applicationDeploymentHookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_deployment_hook"
}

func (r *applicationDeploymentHookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a deployment hook (webhook) for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "The ID of the application. Acts as the resource identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The deployment hook URL. Send an HTTP POST request to it to trigger a deployment.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *applicationDeploymentHookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationDeploymentHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, err := r.client.EnableDeploymentHook(ctx, plan.ApplicationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Deployment Hook", err.Error())
		return
	}

	plan.URL = types.StringValue(hook.URL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, err := r.client.GetDeploymentHook(ctx, state.ApplicationID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application Deployment Hook", err.Error())
		return
	}

	state.URL = types.StringValue(hook.URL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is never called: application_id, the only argument, forces replacement.
func (r *applicationDeploymentHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DisableDeploymentHook(ctx, state.ApplicationID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Application Deployment Hook", err.Error())
	}
}

func (r *applicationDeploymentHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}
