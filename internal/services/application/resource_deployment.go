package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource              = &applicationDeploymentResource{}
	_ resource.ResourceWithConfigure = &applicationDeploymentResource{}
)

type applicationDeploymentResource struct {
	client *client.SevallaClient
}

type ApplicationDeploymentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	IsRestart     types.Bool   `tfsdk:"is_restart"`
	Triggers      types.Map    `tfsdk:"triggers"`
}

func NewDeploymentResource() resource.Resource {
	return &applicationDeploymentResource{}
}

func (r *applicationDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_deployment"
}

func (r *applicationDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a deployment for a Sevalla application. Use the triggers attribute to re-deploy when dependent resources change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application to deploy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_restart": schema.BoolAttribute{
				Description: "When true, redeploys using the existing build artifact, skipping the build step.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"triggers": schema.MapAttribute{
				Description: "A map of arbitrary strings that, when changed, will trigger a new deployment. Use this to link deployments to changes in other resources.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *applicationDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationDeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	triggerReq := &client.TriggerDeploymentRequest{
		IsRestart: plan.IsRestart.ValueBool(),
	}

	deploymentID, err := r.client.TriggerDeployment(ctx, appID, triggerReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Triggering Deployment",
			"Could not trigger deployment for application "+appID+": "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(deploymentID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// Deployment is a fire-and-forget action; state is kept as-is.
}

func (r *applicationDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// is_restart can change in-place. Re-trigger a deployment.
	var plan ApplicationDeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	triggerReq := &client.TriggerDeploymentRequest{
		IsRestart: plan.IsRestart.ValueBool(),
	}

	deploymentID, err := r.client.TriggerDeployment(ctx, appID, triggerReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Triggering Deployment",
			"Could not trigger deployment for application "+appID+": "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(deploymentID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: removing the resource does not cancel or undo a deployment.
}
