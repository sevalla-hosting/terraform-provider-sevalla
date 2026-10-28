package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla/terraform-provider-sevalla/internal/client"
)

// deploymentTimeout bounds how long wait_for_completion waits.
const deploymentTimeout = 60 * time.Minute

var commitSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

var (
	_ resource.Resource                   = &applicationDeploymentResource{}
	_ resource.ResourceWithConfigure      = &applicationDeploymentResource{}
	_ resource.ResourceWithValidateConfig = &applicationDeploymentResource{}
)

type applicationDeploymentResource struct {
	client *client.SevallaClient
}

type ApplicationDeploymentResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ApplicationID     types.String `tfsdk:"application_id"`
	IsRestart         types.Bool   `tfsdk:"is_restart"`
	Branch            types.String `tfsdk:"branch"`
	CommitSHA         types.String `tfsdk:"commit_sha"`
	DockerImage       types.String `tfsdk:"docker_image"`
	SkipBuildCache    types.Bool   `tfsdk:"skip_build_cache"`
	WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
	Triggers          types.Map    `tfsdk:"triggers"`
	Status            types.String `tfsdk:"status"`
}

func NewDeploymentResource() resource.Resource {
	return &applicationDeploymentResource{}
}

func (r *applicationDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_deployment"
}

func (r *applicationDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplaceString := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Triggers a deployment for a Sevalla application. Changing any argument other than `wait_for_completion`, including `triggers`, triggers a new deployment. Destroying this resource does not cancel or roll back the deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description:   "The ID of the application to deploy.",
				Required:      true,
				PlanModifiers: requiresReplaceString,
			},
			"is_restart": schema.BoolAttribute{
				Description: "When true, redeploys the build artifact of the last successful deployment and skips the build step, which is enough to apply runtime changes such as environment variables. Set to `false` to build the latest code. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"branch": schema.StringAttribute{
				Description:   "The Git branch to deploy. Defaults to the application's default branch. Only applies to Git-based applications.",
				Optional:      true,
				PlanModifiers: requiresReplaceString,
			},
			"commit_sha": schema.StringAttribute{
				Description:   "The full 40-character SHA of the commit to deploy. Defaults to the latest commit on the branch. Requires `is_restart = false`.",
				Optional:      true,
				PlanModifiers: requiresReplaceString,
				Validators: []validator.String{
					stringvalidator.RegexMatches(commitSHAPattern, "must be a full 40-character lowercase commit SHA"),
				},
			},
			"docker_image": schema.StringAttribute{
				Description:   "The Docker image reference to deploy (e.g. `nginx:latest`). Only applies to applications with the `dockerImage` source.",
				Optional:      true,
				PlanModifiers: requiresReplaceString,
			},
			"skip_build_cache": schema.BoolAttribute{
				Description: "When true, every layer is rebuilt instead of being reused from the build cache. Defaults to `false`. Requires `is_restart = false`.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_completion": schema.BoolAttribute{
				Description: "When true, apply waits (up to 60 minutes) until the deployment finishes and fails if it fails or is cancelled. Defaults to `false`.",
				Optional:    true,
			},
			"triggers": schema.MapAttribute{
				Description: "A map of arbitrary strings that, when changed, triggers a new deployment. Use this to redeploy when other resources change.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The deployment status: `waiting`, `inProgress`, `success`, `failed`, `cancelled` or `skipped`.",
				Computed:    true,
			},
		},
	}
}

// ValidateConfig rejects options the API cannot combine with a restart,
// which is_restart defaults to.
func (r *applicationDeploymentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ApplicationDeploymentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.IsRestart.IsUnknown() || (!config.IsRestart.IsNull() && !config.IsRestart.ValueBool()) {
		return
	}
	if !config.CommitSHA.IsNull() && !config.CommitSHA.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("commit_sha"), "Invalid Attribute Combination", "commit_sha builds a specific commit, so it requires is_restart = false.")
	}
	if config.SkipBuildCache.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("skip_build_cache"), "Invalid Attribute Combination", "skip_build_cache rebuilds the application, so it requires is_restart = false.")
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
	deployment, err := r.client.TriggerDeployment(ctx, appID, &client.TriggerDeploymentRequest{
		Branch:         plan.Branch.ValueStringPointer(),
		CommitSHA:      plan.CommitSHA.ValueStringPointer(),
		DockerImage:    plan.DockerImage.ValueStringPointer(),
		IsRestart:      plan.IsRestart.ValueBool(),
		SkipBuildCache: plan.SkipBuildCache.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Triggering Deployment",
			"Could not trigger deployment for application "+appID+": "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(deployment.ID)
	plan.Status = optionalString(deployment.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	if !plan.WaitForCompletion.ValueBool() {
		return
	}

	err = client.Poll(ctx, deploymentTimeout, func() (bool, error) {
		d, err := r.client.GetDeployment(ctx, appID, deployment.ID)
		var apiErr *client.APIError
		if err != nil && errors.As(err, &apiErr) && apiErr.StatusCode < 500 {
			return false, err
		}
		if err != nil {
			// Keep waiting through network errors and server errors.
			return false, nil
		}
		plan.Status = optionalString(d.Status)
		switch plan.Status.ValueString() {
		case "success", "skipped":
			return true, nil
		case "failed", "cancelled":
			return false, fmt.Errorf("deployment %s %s", deployment.ID, plan.Status.ValueString())
		}
		return false, nil
	})
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("status"), plan.Status)...)
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for Deployment", err.Error())
	}
}

func (r *applicationDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationDeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.GetDeployment(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		// A deployment that is gone stays in state: recreating it would deploy again.
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error Reading Deployment", "Could not read deployment ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Status = optionalString(deployment.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is only reached when wait_for_completion changes, which needs no API call.
func (r *applicationDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ApplicationDeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: removing the resource does not cancel or undo a deployment.
}
