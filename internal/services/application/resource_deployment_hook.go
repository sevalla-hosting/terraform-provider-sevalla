package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

type deploymentHookResponse struct {
	URL string `json:"url"`
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
				Description: "The deployment hook URL.",
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

	appID := plan.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.client.BaseURL+"/applications/"+appID+"/deployment-hook", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Deployment Hook", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Deployment Hook", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Creating Application Deployment Hook", parseHTTPError(httpResp).Error())
		return
	}

	var hook deploymentHookResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&hook); err != nil {
		resp.Diagnostics.AddError("Error Creating Application Deployment Hook", fmt.Sprintf("decoding response: %s", err))
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

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL+"/applications/"+appID+"/deployment-hook", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application Deployment Hook", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application Deployment Hook", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Reading Application Deployment Hook", parseHTTPError(httpResp).Error())
		return
	}

	var hook deploymentHookResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&hook); err != nil {
		resp.Diagnostics.AddError("Error Reading Application Deployment Hook", fmt.Sprintf("decoding response: %s", err))
		return
	}

	state.URL = types.StringValue(hook.URL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationDeploymentHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update regenerates the webhook URL.
	appID := plan.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, r.client.BaseURL+"/applications/"+appID+"/deployment-hook", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application Deployment Hook", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application Deployment Hook", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Updating Application Deployment Hook", parseHTTPError(httpResp).Error())
		return
	}

	var hook deploymentHookResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&hook); err != nil {
		resp.Diagnostics.AddError("Error Updating Application Deployment Hook", fmt.Sprintf("decoding response: %s", err))
		return
	}

	plan.URL = types.StringValue(hook.URL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDeploymentHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationDeploymentHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.client.BaseURL+"/applications/"+appID+"/deployment-hook", nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Deployment Hook", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Deployment Hook", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Error Deleting Application Deployment Hook", parseHTTPError(httpResp).Error())
		return
	}
}

func (r *applicationDeploymentHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}
