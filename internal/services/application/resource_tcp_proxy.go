package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla/terraform-provider-sevalla/internal/client"
)

// portReadyTimeout bounds how long Create waits for a TCP proxy or private port.
const portReadyTimeout = 10 * time.Minute

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

func NewTCPProxyResource() resource.Resource {
	return &applicationTCPProxyResource{}
}

func (r *applicationTCPProxyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_tcp_proxy"
}

func (r *applicationTCPProxyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TCP proxy that exposes a port of a Sevalla application process to the internet.",
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
				Description: "The ID of the process to attach the TCP proxy to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The internal container port to expose (1-65535).",
				Required:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 65535)},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"external_port": schema.Int64Attribute{
				Description: "The public port to connect to. It can change once when the proxy moves behind the external proxy, so read it together with `hostname`.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The public hostname to connect to.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the TCP proxy: `creating`, `ready`, `failed` or `deleting`.",
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
	proxy, err := r.client.CreateTCPProxy(ctx, appID, &client.CreatePortRequest{
		ProcessID: plan.ProcessID.ValueString(),
		Port:      plan.Port.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating TCP Proxy", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), appID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), proxy.ID)...)

	err = client.Poll(ctx, portReadyTimeout, func() (bool, error) {
		found, err := r.find(ctx, appID, proxy.ID)
		if err != nil || found == nil {
			return false, err
		}
		proxy = found
		switch proxy.Status {
		case "ready":
			return true, nil
		case "failed", "deleting":
			return false, fmt.Errorf("TCP proxy %s is %s", proxy.ID, proxy.Status)
		}
		return false, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for TCP Proxy", err.Error())
		return
	}

	flattenTCPProxy(proxy, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationTCPProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationTCPProxyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy, err := r.find(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading TCP Proxy", err.Error())
		return
	}
	if proxy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenTCPProxy(proxy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// find returns the TCP proxy with the given ID, or nil when there is none.
// The API has no endpoint to get a single TCP proxy.
func (r *applicationTCPProxyResource) find(ctx context.Context, appID, proxyID string) (*client.TCPProxy, error) {
	proxies, err := r.client.ListTCPProxies(ctx, appID)
	if err != nil {
		return nil, err
	}
	for i := range proxies {
		if proxies[i].ID == proxyID {
			return &proxies[i], nil
		}
	}
	return nil, nil
}

func (r *applicationTCPProxyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "TCP proxies cannot be updated. Change a required attribute to replace the resource.")
}

func (r *applicationTCPProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationTCPProxyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTCPProxy(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting TCP Proxy", err.Error())
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

func flattenTCPProxy(p *client.TCPProxy, model *ApplicationTCPProxyResourceModel) {
	model.ID = types.StringValue(p.ID)
	model.ProcessID = types.StringValue(p.ProcessID)
	model.Port = types.Int64Value(p.Port)
	model.ExternalPort = types.Int64Value(p.ExternalPort)
	model.Hostname = types.StringValue(p.Hostname)
	model.Status = types.StringValue(p.Status)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)
}
