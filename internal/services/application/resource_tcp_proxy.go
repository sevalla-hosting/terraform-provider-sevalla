package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationTCPProxyResource{}
	_ resource.ResourceWithConfigure   = &applicationTCPProxyResource{}
	_ resource.ResourceWithImportState = &applicationTCPProxyResource{}
)

type applicationTCPProxyResource struct {
	client *client.SevallaClient
}

type ApplicationTCPProxyResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	ProcessID     types.String `tfsdk:"process_id"`
	Port          types.Int64  `tfsdk:"port"`
	ExternalPort  types.Int64  `tfsdk:"external_port"`
	Hostname      types.String `tfsdk:"hostname"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type createTCPProxyRequest struct {
	ProcessID string `json:"process_id"`
	Port      int64  `json:"port"`
}

type tcpProxyResponse struct {
	ID           string  `json:"id"`
	ProcessID    *string `json:"process_id"`
	Port         int64   `json:"port"`
	ExternalPort int64   `json:"external_port"`
	Hostname     *string `json:"hostname"`
	Status       *string `json:"status"`
	CreatedAt    *string `json:"created_at"`
	UpdatedAt    *string `json:"updated_at"`
}

func NewTCPProxyResource() resource.Resource {
	return &applicationTCPProxyResource{}
}

func (r *applicationTCPProxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_tcp_proxy"
}

func (r *applicationTCPProxyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCP proxy for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the TCP proxy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this TCP proxy belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"process_id": schema.StringAttribute{
				Description: "The ID of the process this TCP proxy belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The internal port number.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"external_port": schema.Int64Attribute{
				Description: "The external port number assigned by the platform.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname for the TCP proxy.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the TCP proxy.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the TCP proxy was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the TCP proxy was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *applicationTCPProxyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationTCPProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationTCPProxyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	createReq := createTCPProxyRequest{
		ProcessID: plan.ProcessID.ValueString(),
		Port:      plan.Port.ValueInt64(),
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application TCP Proxy", err.Error())
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.client.BaseURL+"/applications/"+appID+"/tcp-proxies", bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application TCP Proxy", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application TCP Proxy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Creating Application TCP Proxy", parseHTTPError(httpResp).Error())
		return
	}

	var proxy tcpProxyResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&proxy); err != nil {
		resp.Diagnostics.AddError("Error Creating Application TCP Proxy", fmt.Sprintf("decoding response: %s", err))
		return
	}

	flattenTCPProxy(&proxy, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationTCPProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationTCPProxyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL+"/applications/"+appID+"/tcp-proxies/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application TCP Proxy", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application TCP Proxy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Reading Application TCP Proxy", parseHTTPError(httpResp).Error())
		return
	}

	var proxy tcpProxyResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&proxy); err != nil {
		resp.Diagnostics.AddError("Error Reading Application TCP Proxy", fmt.Sprintf("decoding response: %s", err))
		return
	}

	flattenTCPProxy(&proxy, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationTCPProxyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Application TCP proxies cannot be updated. Delete and recreate the resource instead.",
	)
}

func (r *applicationTCPProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationTCPProxyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.client.BaseURL+"/applications/"+appID+"/tcp-proxies/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application TCP Proxy", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application TCP Proxy", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Error Deleting Application TCP Proxy", parseHTTPError(httpResp).Error())
		return
	}
}

func (r *applicationTCPProxyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'application_id/tcp_proxy_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenTCPProxy(proxy *tcpProxyResponse, model *ApplicationTCPProxyResourceModel) {
	model.ID = types.StringValue(proxy.ID)
	if proxy.ProcessID != nil {
		model.ProcessID = types.StringValue(*proxy.ProcessID)
	}
	model.Port = types.Int64Value(proxy.Port)
	model.ExternalPort = types.Int64Value(proxy.ExternalPort)
	if proxy.Hostname != nil {
		model.Hostname = types.StringValue(*proxy.Hostname)
	} else {
		model.Hostname = types.StringNull()
	}
	if proxy.Status != nil {
		model.Status = types.StringValue(*proxy.Status)
	} else {
		model.Status = types.StringNull()
	}
	if proxy.CreatedAt != nil {
		model.CreatedAt = types.StringValue(*proxy.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
	if proxy.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(*proxy.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}
}
