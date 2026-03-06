package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationProcessResource{}
	_ resource.ResourceWithConfigure   = &applicationProcessResource{}
	_ resource.ResourceWithImportState = &applicationProcessResource{}
)

type applicationProcessResource struct {
	client *client.SevallaClient
}

type ApplicationProcessResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ApplicationID    types.String `tfsdk:"application_id"`
	DisplayName      types.String `tfsdk:"display_name"`
	Type             types.String `tfsdk:"type"`
	Entrypoint       types.String `tfsdk:"entrypoint"`
	Port             types.Int64  `tfsdk:"port"`
	ResourceTypeID   types.String `tfsdk:"resource_type_id"`
	ScalingStrategy  types.Object `tfsdk:"scaling_strategy"`
	Schedule         types.String `tfsdk:"schedule"`
	TimeZone         types.String `tfsdk:"time_zone"`
	JobStartPolicy   types.String `tfsdk:"job_start_policy"`
	IsIngressEnabled types.Bool   `tfsdk:"is_ingress_enabled"`
	IngressProtocol  types.String `tfsdk:"ingress_protocol"`
	LivenessProbe    types.Object `tfsdk:"liveness_probe"`
	ReadinessProbe   types.Object `tfsdk:"readiness_probe"`
	Key              types.String `tfsdk:"key"`
	ResourceTypeName types.String `tfsdk:"resource_type_name"`
	CpuLimit         types.Int64  `tfsdk:"cpu_limit"`
	MemoryLimit      types.Int64  `tfsdk:"memory_limit"`
	InternalHostname types.String `tfsdk:"internal_hostname"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// Attr type maps for nested objects.
var scalingStrategyAttrTypes = map[string]attr.Type{
	"type":                        types.StringType,
	"instance_count":              types.Int64Type,
	"min_instance_count":          types.Int64Type,
	"max_instance_count":          types.Int64Type,
	"target_cpu_percent":          types.Int64Type,
	"target_memory_percent":       types.Int64Type,
	"scale_up_interval_seconds":   types.Int64Type,
	"scale_up_increment":          types.Int64Type,
	"scale_down_interval_seconds": types.Int64Type,
	"scale_down_increment":        types.Int64Type,
}

var httpHeaderAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

var probeExecAttrTypes = map[string]attr.Type{
	"command": types.ListType{ElemType: types.StringType},
}

var probeHttpGetAttrTypes = map[string]attr.Type{
	"path":         types.StringType,
	"port":         types.Int64Type,
	"host":         types.StringType,
	"scheme":       types.StringType,
	"http_headers": types.ListType{ElemType: types.ObjectType{AttrTypes: httpHeaderAttrTypes}},
}

var probeTcpSocketAttrTypes = map[string]attr.Type{
	"host": types.StringType,
	"port": types.Int64Type,
}

var probeAttrTypes = map[string]attr.Type{
	"exec":                  types.ObjectType{AttrTypes: probeExecAttrTypes},
	"http_get":              types.ObjectType{AttrTypes: probeHttpGetAttrTypes},
	"tcp_socket":            types.ObjectType{AttrTypes: probeTcpSocketAttrTypes},
	"initial_delay_seconds": types.Int64Type,
	"period_seconds":        types.Int64Type,
	"timeout_seconds":       types.Int64Type,
	"success_threshold":     types.Int64Type,
	"failure_threshold":     types.Int64Type,
}

// API request/response types.

type createProcessRequest struct {
	DisplayName     string                  `json:"display_name"`
	Type            string                  `json:"type"`
	Entrypoint      string                  `json:"entrypoint,omitempty"`
	Port            *int64                  `json:"port,omitempty"`
	ResourceTypeID  string                  `json:"resource_type_id"`
	ScalingStrategy *scalingStrategyRequest `json:"scaling_strategy"`
	Schedule        *string                 `json:"schedule,omitempty"`
	TimeZone        *string                 `json:"time_zone,omitempty"`
	JobStartPolicy  *string                 `json:"job_start_policy,omitempty"`
	LivenessProbe   *probeRequest           `json:"liveness_probe,omitempty"`
	ReadinessProbe  *probeRequest           `json:"readiness_probe,omitempty"`
}

type updateProcessRequest struct {
	DisplayName      *string                 `json:"display_name,omitempty"`
	Entrypoint       *string                 `json:"entrypoint,omitempty"`
	Port             *int64                  `json:"port,omitempty"`
	ResourceTypeID   *string                 `json:"resource_type_id,omitempty"`
	ScalingStrategy  *scalingStrategyRequest `json:"scaling_strategy,omitempty"`
	Schedule         *string                 `json:"schedule,omitempty"`
	TimeZone         *string                 `json:"time_zone,omitempty"`
	JobStartPolicy   *string                 `json:"job_start_policy,omitempty"`
	IsIngressEnabled *bool                   `json:"is_ingress_enabled,omitempty"`
	IngressProtocol  *string                 `json:"ingress_protocol,omitempty"`
	LivenessProbe    *probeRequest           `json:"liveness_probe,omitempty"`
	ReadinessProbe   *probeRequest           `json:"readiness_probe,omitempty"`
}

type scalingStrategyRequest struct {
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type manualScalingConfig struct {
	InstanceCount int64 `json:"instanceCount"`
}

type horizontalScalingConfig struct {
	MinInstanceCount         int64  `json:"minInstanceCount"`
	MaxInstanceCount         int64  `json:"maxInstanceCount"`
	TargetCpuPercent         *int64 `json:"targetCpuPercent,omitempty"`
	TargetMemoryPercent      *int64 `json:"targetMemoryPercent,omitempty"`
	ScaleUpIntervalSeconds   *int64 `json:"scaleUpIntervalSeconds,omitempty"`
	ScaleUpIncrement         *int64 `json:"scaleUpIncrement,omitempty"`
	ScaleDownIntervalSeconds *int64 `json:"scaleDownIntervalSeconds,omitempty"`
	ScaleDownIncrement       *int64 `json:"scaleDownIncrement,omitempty"`
}

type probeRequest struct {
	Exec                *probeExecRequest    `json:"exec,omitempty"`
	HttpGet             *probeHttpGetRequest `json:"httpGet,omitempty"`
	TcpSocket           *probeTcpSocketRequest `json:"tcpSocket,omitempty"`
	InitialDelaySeconds *int64               `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       *int64               `json:"periodSeconds,omitempty"`
	TimeoutSeconds      *int64               `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    *int64               `json:"successThreshold,omitempty"`
	FailureThreshold    *int64               `json:"failureThreshold,omitempty"`
}

type probeExecRequest struct {
	Command []string `json:"command"`
}

type probeHttpGetRequest struct {
	Path        *string             `json:"path,omitempty"`
	Port        *int64              `json:"port,omitempty"`
	Host        *string             `json:"host,omitempty"`
	Scheme      *string             `json:"scheme,omitempty"`
	HttpHeaders []httpHeaderRequest `json:"httpHeaders,omitempty"`
}

type httpHeaderRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type probeTcpSocketRequest struct {
	Host *string `json:"host,omitempty"`
	Port int64   `json:"port"`
}

type processResponse struct {
	ID               string                   `json:"id"`
	AppID            *string                  `json:"app_id"`
	Key              string                   `json:"key"`
	Type             string                   `json:"type"`
	DisplayName      *string                  `json:"display_name"`
	Entrypoint       string                   `json:"entrypoint"`
	Port             *int64                   `json:"port"`
	IsIngressEnabled bool                     `json:"is_ingress_enabled"`
	IngressProtocol  *string                  `json:"ingress_protocol"`
	ScalingStrategy  *scalingStrategyResponse `json:"scaling_strategy"`
	ResourceTypeID   *string                  `json:"resource_type_id"`
	ResourceTypeName *string                  `json:"resource_type_name"`
	CpuLimit         *int64                   `json:"cpu_limit"`
	MemoryLimit      *int64                   `json:"memory_limit"`
	Schedule         *string                  `json:"schedule"`
	TimeZone         *string                  `json:"time_zone"`
	JobStartPolicy   *string                  `json:"job_start_policy"`
	LivenessProbe    *probeResponse           `json:"liveness_probe"`
	ReadinessProbe   *probeResponse           `json:"readiness_probe"`
	InternalHostname *string                  `json:"internal_hostname"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

type scalingStrategyResponse struct {
	Type   string              `json:"type"`
	Config scalingConfigResponse `json:"config"`
}

type scalingConfigResponse struct {
	InstanceCount            *int64 `json:"instanceCount,omitempty"`
	MinInstanceCount         *int64 `json:"minInstanceCount,omitempty"`
	MaxInstanceCount         *int64 `json:"maxInstanceCount,omitempty"`
	TargetCpuPercent         *int64 `json:"targetCpuPercent,omitempty"`
	TargetMemoryPercent      *int64 `json:"targetMemoryPercent,omitempty"`
	ScaleUpIntervalSeconds   *int64 `json:"scaleUpIntervalSeconds,omitempty"`
	ScaleUpIncrement         *int64 `json:"scaleUpIncrement,omitempty"`
	ScaleDownIntervalSeconds *int64 `json:"scaleDownIntervalSeconds,omitempty"`
	ScaleDownIncrement       *int64 `json:"scaleDownIncrement,omitempty"`
}

type probeResponse struct {
	Exec                *probeExecResponse    `json:"exec,omitempty"`
	HttpGet             *probeHttpGetResponse `json:"httpGet,omitempty"`
	TcpSocket           *probeTcpSocketResponse `json:"tcpSocket,omitempty"`
	InitialDelaySeconds *int64                `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       *int64                `json:"periodSeconds,omitempty"`
	TimeoutSeconds      *int64                `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    *int64                `json:"successThreshold,omitempty"`
	FailureThreshold    *int64                `json:"failureThreshold,omitempty"`
}

type probeExecResponse struct {
	Command []string `json:"command"`
}

type probeHttpGetResponse struct {
	Path        *string              `json:"path,omitempty"`
	Port        *int64               `json:"port,omitempty"`
	Host        *string              `json:"host,omitempty"`
	Scheme      *string              `json:"scheme,omitempty"`
	HttpHeaders []httpHeaderResponse `json:"httpHeaders,omitempty"`
}

type httpHeaderResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type probeTcpSocketResponse struct {
	Host *string `json:"host,omitempty"`
	Port int64   `json:"port"`
}

func NewProcessResource() resource.Resource {
	return &applicationProcessResource{}
}

func (r *applicationProcessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_process"
}

func (r *applicationProcessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a process for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the process.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this process belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name for the process (1-255 characters).",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The process type. Valid values: web, worker, cron, job.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entrypoint": schema.StringAttribute{
				Description: "Command used to start the process (max 1024 characters).",
				Optional:    true,
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number the process listens on (1-65535).",
				Optional:    true,
				Computed:    true,
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The resource type ID determining CPU and memory allocation.",
				Required:    true,
			},
			"scaling_strategy": schema.SingleNestedAttribute{
				Description: "The scaling configuration for the process.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "The scaling type. Valid values: manual, horizontal.",
						Required:    true,
					},
					"instance_count": schema.Int64Attribute{
						Description: "The number of instances (manual scaling only). Range: 0-50.",
						Optional:    true,
					},
					"min_instance_count": schema.Int64Attribute{
						Description: "Minimum number of instances (horizontal scaling only). Range: 0-50.",
						Optional:    true,
					},
					"max_instance_count": schema.Int64Attribute{
						Description: "Maximum number of instances (horizontal scaling only). Range: 0-50.",
						Optional:    true,
					},
					"target_cpu_percent": schema.Int64Attribute{
						Description: "Target CPU utilization percentage for autoscaling (1-100).",
						Optional:    true,
					},
					"target_memory_percent": schema.Int64Attribute{
						Description: "Target memory utilization percentage for autoscaling (1-100).",
						Optional:    true,
					},
					"scale_up_interval_seconds": schema.Int64Attribute{
						Description: "Seconds between scale-up evaluations (1-60000).",
						Optional:    true,
					},
					"scale_up_increment": schema.Int64Attribute{
						Description: "Number of instances to add per scale-up (1-5).",
						Optional:    true,
					},
					"scale_down_interval_seconds": schema.Int64Attribute{
						Description: "Seconds between scale-down evaluations (1-60000).",
						Optional:    true,
					},
					"scale_down_increment": schema.Int64Attribute{
						Description: "Number of instances to remove per scale-down (1-5).",
						Optional:    true,
					},
				},
			},
			"schedule": schema.StringAttribute{
				Description: "Cron schedule expression (required for cron processes, max 100 characters).",
				Optional:    true,
			},
			"time_zone": schema.StringAttribute{
				Description: "IANA time zone for cron schedule (e.g. America/New_York).",
				Optional:    true,
				Computed:    true,
			},
			"job_start_policy": schema.StringAttribute{
				Description: "When the job runs relative to deployments (job processes only). Valid values: beforeDeployment, afterSuccessDeployment, afterFailedDeployment.",
				Optional:    true,
			},
			"is_ingress_enabled": schema.BoolAttribute{
				Description: "Whether external traffic routing is enabled.",
				Optional:    true,
				Computed:    true,
			},
			"ingress_protocol": schema.StringAttribute{
				Description: "The ingress protocol. Valid values: http, grpc.",
				Optional:    true,
				Computed:    true,
			},
			"liveness_probe": schema.SingleNestedAttribute{
				Description: "Liveness probe configuration (web processes only).",
				Optional:    true,
				Computed:    true,
				Attributes:  probeSchemaAttributes(),
			},
			"readiness_probe": schema.SingleNestedAttribute{
				Description: "Readiness probe configuration (web processes only).",
				Optional:    true,
				Computed:    true,
				Attributes:  probeSchemaAttributes(),
			},
			"key": schema.StringAttribute{
				Description: "Unique process key within the application.",
				Computed:    true,
			},
			"resource_type_name": schema.StringAttribute{
				Description: "The name of the resource type (machine size).",
				Computed:    true,
			},
			"cpu_limit": schema.Int64Attribute{
				Description: "CPU allocation in millicores.",
				Computed:    true,
			},
			"memory_limit": schema.Int64Attribute{
				Description: "Memory allocation in megabytes.",
				Computed:    true,
			},
			"internal_hostname": schema.StringAttribute{
				Description: "Internal Kubernetes DNS hostname for service-to-service communication.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the process was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the process was last updated.",
				Computed:    true,
			},
		},
	}
}

func probeSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"exec": schema.SingleNestedAttribute{
			Description: "Exec probe: runs a command inside the container.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"command": schema.ListAttribute{
					Description: "Command to execute.",
					Required:    true,
					ElementType: types.StringType,
				},
			},
		},
		"http_get": schema.SingleNestedAttribute{
			Description: "HTTP GET probe.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"path": schema.StringAttribute{
					Description: "Path to probe.",
					Optional:    true,
				},
				"port": schema.Int64Attribute{
					Description: "Port to probe (1-65535).",
					Optional:    true,
				},
				"host": schema.StringAttribute{
					Description: "Hostname for the probe request.",
					Optional:    true,
				},
				"scheme": schema.StringAttribute{
					Description: "URI scheme. Valid values: HTTP, HTTPS.",
					Optional:    true,
				},
				"http_headers": schema.ListNestedAttribute{
					Description: "Custom HTTP headers for the probe request.",
					Optional:    true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Description: "Header name.",
								Required:    true,
							},
							"value": schema.StringAttribute{
								Description: "Header value.",
								Required:    true,
							},
						},
					},
				},
			},
		},
		"tcp_socket": schema.SingleNestedAttribute{
			Description: "TCP socket probe.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"host": schema.StringAttribute{
					Description: "Hostname to connect to.",
					Optional:    true,
				},
				"port": schema.Int64Attribute{
					Description: "Port to connect to (1-65535).",
					Required:    true,
				},
			},
		},
		"initial_delay_seconds": schema.Int64Attribute{
			Description: "Seconds before the probe starts after container start.",
			Optional:    true,
		},
		"period_seconds": schema.Int64Attribute{
			Description: "How often (in seconds) to perform the probe.",
			Optional:    true,
		},
		"timeout_seconds": schema.Int64Attribute{
			Description: "Seconds after which the probe times out.",
			Optional:    true,
		},
		"success_threshold": schema.Int64Attribute{
			Description: "Consecutive successes required to be considered healthy.",
			Optional:    true,
		},
		"failure_threshold": schema.Int64Attribute{
			Description: "Consecutive failures required to be considered unhealthy.",
			Optional:    true,
		},
	}
}

func (r *applicationProcessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationProcessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationProcessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateProcessRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	process, err := r.doCreateProcess(ctx, plan.ApplicationID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Process", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenProcess(ctx, process, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationProcessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationProcessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	process, err := r.doGetProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application Process", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenProcess(ctx, process, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationProcessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationProcessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ApplicationProcessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateProcessRequest(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	process, err := r.doUpdateProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application Process", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenProcess(ctx, process, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationProcessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationProcessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.doDeleteProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Process", err.Error())
		return
	}
}

func (r *applicationProcessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'application_id/process_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// Build helpers.

func buildCreateProcessRequest(ctx context.Context, plan *ApplicationProcessResourceModel) (*createProcessRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &createProcessRequest{
		DisplayName:    plan.DisplayName.ValueString(),
		Type:           plan.Type.ValueString(),
		ResourceTypeID: plan.ResourceTypeID.ValueString(),
	}

	if !plan.Entrypoint.IsNull() && !plan.Entrypoint.IsUnknown() {
		req.Entrypoint = plan.Entrypoint.ValueString()
	}
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		v := plan.Port.ValueInt64()
		req.Port = &v
	}
	if !plan.Schedule.IsNull() && !plan.Schedule.IsUnknown() {
		v := plan.Schedule.ValueString()
		req.Schedule = &v
	}
	if !plan.TimeZone.IsNull() && !plan.TimeZone.IsUnknown() {
		v := plan.TimeZone.ValueString()
		req.TimeZone = &v
	}
	if !plan.JobStartPolicy.IsNull() && !plan.JobStartPolicy.IsUnknown() {
		v := plan.JobStartPolicy.ValueString()
		req.JobStartPolicy = &v
	}
	// Note: is_ingress_enabled and ingress_protocol are NOT accepted on create
	// (not in the POST request schema). They can only be set via PATCH update.

	ss, d := buildScalingStrategyRequest(ctx, plan.ScalingStrategy)
	diags.Append(d...)
	if !diags.HasError() {
		req.ScalingStrategy = ss
	}

	if !plan.LivenessProbe.IsNull() && !plan.LivenessProbe.IsUnknown() {
		p, d := buildProbeRequest(ctx, plan.LivenessProbe)
		diags.Append(d...)
		req.LivenessProbe = p
	}
	if !plan.ReadinessProbe.IsNull() && !plan.ReadinessProbe.IsUnknown() {
		p, d := buildProbeRequest(ctx, plan.ReadinessProbe)
		diags.Append(d...)
		req.ReadinessProbe = p
	}

	return req, diags
}

func buildUpdateProcessRequest(ctx context.Context, plan, state *ApplicationProcessResourceModel) (*updateProcessRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &updateProcessRequest{}

	if !plan.DisplayName.Equal(state.DisplayName) {
		v := plan.DisplayName.ValueString()
		req.DisplayName = &v
	}
	if !plan.Entrypoint.Equal(state.Entrypoint) {
		v := plan.Entrypoint.ValueString()
		req.Entrypoint = &v
	}
	if !plan.Port.Equal(state.Port) {
		v := plan.Port.ValueInt64()
		req.Port = &v
	}
	if !plan.ResourceTypeID.Equal(state.ResourceTypeID) {
		v := plan.ResourceTypeID.ValueString()
		req.ResourceTypeID = &v
	}
	if !plan.Schedule.Equal(state.Schedule) {
		if plan.Schedule.IsNull() {
			empty := ""
			req.Schedule = &empty
		} else {
			v := plan.Schedule.ValueString()
			req.Schedule = &v
		}
	}
	if !plan.TimeZone.Equal(state.TimeZone) {
		if plan.TimeZone.IsNull() {
			empty := ""
			req.TimeZone = &empty
		} else {
			v := plan.TimeZone.ValueString()
			req.TimeZone = &v
		}
	}
	if !plan.JobStartPolicy.Equal(state.JobStartPolicy) {
		v := plan.JobStartPolicy.ValueString()
		req.JobStartPolicy = &v
	}
	if !plan.IsIngressEnabled.Equal(state.IsIngressEnabled) {
		v := plan.IsIngressEnabled.ValueBool()
		req.IsIngressEnabled = &v
	}
	if !plan.IngressProtocol.Equal(state.IngressProtocol) {
		v := plan.IngressProtocol.ValueString()
		req.IngressProtocol = &v
	}
	if !plan.ScalingStrategy.Equal(state.ScalingStrategy) {
		ss, d := buildScalingStrategyRequest(ctx, plan.ScalingStrategy)
		diags.Append(d...)
		req.ScalingStrategy = ss
	}
	if !plan.LivenessProbe.Equal(state.LivenessProbe) {
		if plan.LivenessProbe.IsNull() {
			req.LivenessProbe = nil
		} else {
			p, d := buildProbeRequest(ctx, plan.LivenessProbe)
			diags.Append(d...)
			req.LivenessProbe = p
		}
	}
	if !plan.ReadinessProbe.Equal(state.ReadinessProbe) {
		if plan.ReadinessProbe.IsNull() {
			req.ReadinessProbe = nil
		} else {
			p, d := buildProbeRequest(ctx, plan.ReadinessProbe)
			diags.Append(d...)
			req.ReadinessProbe = p
		}
	}

	return req, diags
}

func buildScalingStrategyRequest(ctx context.Context, obj types.Object) (*scalingStrategyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	attrs := obj.Attributes()
	stratType := attrs["type"].(types.String).ValueString()

	req := &scalingStrategyRequest{Type: stratType}

	switch stratType {
	case "manual":
		instanceCount := attrs["instance_count"].(types.Int64)
		if instanceCount.IsNull() || instanceCount.IsUnknown() {
			diags.AddError("Invalid Scaling Strategy", "instance_count is required for manual scaling.")
			return nil, diags
		}
		req.Config = manualScalingConfig{
			InstanceCount: instanceCount.ValueInt64(),
		}
	case "horizontal":
		minCount := attrs["min_instance_count"].(types.Int64)
		maxCount := attrs["max_instance_count"].(types.Int64)
		if minCount.IsNull() || minCount.IsUnknown() || maxCount.IsNull() || maxCount.IsUnknown() {
			diags.AddError("Invalid Scaling Strategy", "min_instance_count and max_instance_count are required for horizontal scaling.")
			return nil, diags
		}
		cfg := horizontalScalingConfig{
			MinInstanceCount: minCount.ValueInt64(),
			MaxInstanceCount: maxCount.ValueInt64(),
		}
		if v := attrs["target_cpu_percent"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.TargetCpuPercent = &val
		}
		if v := attrs["target_memory_percent"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.TargetMemoryPercent = &val
		}
		if v := attrs["scale_up_interval_seconds"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.ScaleUpIntervalSeconds = &val
		}
		if v := attrs["scale_up_increment"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.ScaleUpIncrement = &val
		}
		if v := attrs["scale_down_interval_seconds"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.ScaleDownIntervalSeconds = &val
		}
		if v := attrs["scale_down_increment"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			cfg.ScaleDownIncrement = &val
		}
		req.Config = cfg
	default:
		diags.AddError("Invalid Scaling Strategy", fmt.Sprintf("Unknown scaling type: %s. Valid values: manual, horizontal.", stratType))
	}

	return req, diags
}

func buildProbeRequest(ctx context.Context, obj types.Object) (*probeRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	attrs := obj.Attributes()
	req := &probeRequest{}

	// Exec probe
	if execObj, ok := attrs["exec"].(types.Object); ok && !execObj.IsNull() && !execObj.IsUnknown() {
		execAttrs := execObj.Attributes()
		cmdList := execAttrs["command"].(types.List)
		var commands []string
		diags.Append(cmdList.ElementsAs(ctx, &commands, false)...)
		req.Exec = &probeExecRequest{Command: commands}
	}

	// HTTP GET probe
	if httpObj, ok := attrs["http_get"].(types.Object); ok && !httpObj.IsNull() && !httpObj.IsUnknown() {
		httpAttrs := httpObj.Attributes()
		httpReq := &probeHttpGetRequest{}
		if v := httpAttrs["path"].(types.String); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueString()
			httpReq.Path = &val
		}
		if v := httpAttrs["port"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueInt64()
			httpReq.Port = &val
		}
		if v := httpAttrs["host"].(types.String); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueString()
			httpReq.Host = &val
		}
		if v := httpAttrs["scheme"].(types.String); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueString()
			httpReq.Scheme = &val
		}
		if headersList, ok := httpAttrs["http_headers"].(types.List); ok && !headersList.IsNull() && !headersList.IsUnknown() {
			var headerObjs []types.Object
			diags.Append(headersList.ElementsAs(ctx, &headerObjs, false)...)
			for _, h := range headerObjs {
				ha := h.Attributes()
				httpReq.HttpHeaders = append(httpReq.HttpHeaders, httpHeaderRequest{
					Name:  ha["name"].(types.String).ValueString(),
					Value: ha["value"].(types.String).ValueString(),
				})
			}
		}
		req.HttpGet = httpReq
	}

	// TCP socket probe
	if tcpObj, ok := attrs["tcp_socket"].(types.Object); ok && !tcpObj.IsNull() && !tcpObj.IsUnknown() {
		tcpAttrs := tcpObj.Attributes()
		tcpReq := &probeTcpSocketRequest{
			Port: tcpAttrs["port"].(types.Int64).ValueInt64(),
		}
		if v := tcpAttrs["host"].(types.String); !v.IsNull() && !v.IsUnknown() {
			val := v.ValueString()
			tcpReq.Host = &val
		}
		req.TcpSocket = tcpReq
	}

	// Timing
	if v := attrs["initial_delay_seconds"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		val := v.ValueInt64()
		req.InitialDelaySeconds = &val
	}
	if v := attrs["period_seconds"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		val := v.ValueInt64()
		req.PeriodSeconds = &val
	}
	if v := attrs["timeout_seconds"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		val := v.ValueInt64()
		req.TimeoutSeconds = &val
	}
	if v := attrs["success_threshold"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		val := v.ValueInt64()
		req.SuccessThreshold = &val
	}
	if v := attrs["failure_threshold"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		val := v.ValueInt64()
		req.FailureThreshold = &val
	}

	return req, diags
}

// Flatten helpers.

func flattenProcess(ctx context.Context, p *processResponse, model *ApplicationProcessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(p.ID)
	model.Key = types.StringValue(p.Key)
	model.Type = types.StringValue(p.Type)
	model.Entrypoint = types.StringValue(p.Entrypoint)
	model.IsIngressEnabled = types.BoolValue(p.IsIngressEnabled)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)

	if p.DisplayName != nil {
		model.DisplayName = types.StringValue(*p.DisplayName)
	}
	if p.Port != nil {
		model.Port = types.Int64Value(*p.Port)
	} else {
		model.Port = types.Int64Null()
	}
	if p.ResourceTypeID != nil {
		model.ResourceTypeID = types.StringValue(*p.ResourceTypeID)
	}
	if p.ResourceTypeName != nil {
		model.ResourceTypeName = types.StringValue(*p.ResourceTypeName)
	} else {
		model.ResourceTypeName = types.StringNull()
	}
	if p.CpuLimit != nil {
		model.CpuLimit = types.Int64Value(*p.CpuLimit)
	} else {
		model.CpuLimit = types.Int64Null()
	}
	if p.MemoryLimit != nil {
		model.MemoryLimit = types.Int64Value(*p.MemoryLimit)
	} else {
		model.MemoryLimit = types.Int64Null()
	}
	if p.IngressProtocol != nil {
		model.IngressProtocol = types.StringValue(*p.IngressProtocol)
	} else {
		model.IngressProtocol = types.StringNull()
	}
	if p.Schedule != nil && *p.Schedule != "" {
		model.Schedule = types.StringValue(*p.Schedule)
	} else {
		model.Schedule = types.StringNull()
	}
	if p.TimeZone != nil && *p.TimeZone != "" {
		model.TimeZone = types.StringValue(*p.TimeZone)
	} else {
		model.TimeZone = types.StringNull()
	}
	if p.JobStartPolicy != nil && *p.JobStartPolicy != "" {
		model.JobStartPolicy = types.StringValue(*p.JobStartPolicy)
	} else {
		model.JobStartPolicy = types.StringNull()
	}
	if p.InternalHostname != nil {
		model.InternalHostname = types.StringValue(*p.InternalHostname)
	} else {
		model.InternalHostname = types.StringNull()
	}

	// Scaling strategy
	if p.ScalingStrategy != nil {
		ssAttrs := map[string]attr.Value{
			"type": types.StringValue(p.ScalingStrategy.Type),
		}
		cfg := p.ScalingStrategy.Config
		setOptionalInt64 := func(key string, val *int64) {
			if val != nil {
				ssAttrs[key] = types.Int64Value(*val)
			} else {
				ssAttrs[key] = types.Int64Null()
			}
		}
		setOptionalInt64("instance_count", cfg.InstanceCount)
		setOptionalInt64("min_instance_count", cfg.MinInstanceCount)
		setOptionalInt64("max_instance_count", cfg.MaxInstanceCount)
		setOptionalInt64("target_cpu_percent", cfg.TargetCpuPercent)
		setOptionalInt64("target_memory_percent", cfg.TargetMemoryPercent)
		setOptionalInt64("scale_up_interval_seconds", cfg.ScaleUpIntervalSeconds)
		setOptionalInt64("scale_up_increment", cfg.ScaleUpIncrement)
		setOptionalInt64("scale_down_interval_seconds", cfg.ScaleDownIntervalSeconds)
		setOptionalInt64("scale_down_increment", cfg.ScaleDownIncrement)

		obj, d := types.ObjectValue(scalingStrategyAttrTypes, ssAttrs)
		diags.Append(d...)
		model.ScalingStrategy = obj
	} else {
		model.ScalingStrategy = types.ObjectNull(scalingStrategyAttrTypes)
	}

	// Probes
	lp, d := flattenProbe(ctx, p.LivenessProbe)
	diags.Append(d...)
	model.LivenessProbe = lp

	rp, d := flattenProbe(ctx, p.ReadinessProbe)
	diags.Append(d...)
	model.ReadinessProbe = rp

	return diags
}

func flattenProbe(_ context.Context, p *probeResponse) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if p == nil {
		return types.ObjectNull(probeAttrTypes), diags
	}

	attrs := map[string]attr.Value{}

	// Exec
	if p.Exec != nil {
		cmdVals := make([]attr.Value, len(p.Exec.Command))
		for i, c := range p.Exec.Command {
			cmdVals[i] = types.StringValue(c)
		}
		cmdList, d := types.ListValue(types.StringType, cmdVals)
		diags.Append(d...)
		execObj, d := types.ObjectValue(probeExecAttrTypes, map[string]attr.Value{
			"command": cmdList,
		})
		diags.Append(d...)
		attrs["exec"] = execObj
	} else {
		attrs["exec"] = types.ObjectNull(probeExecAttrTypes)
	}

	// HTTP GET
	if p.HttpGet != nil {
		httpAttrs := map[string]attr.Value{}
		if p.HttpGet.Path != nil {
			httpAttrs["path"] = types.StringValue(*p.HttpGet.Path)
		} else {
			httpAttrs["path"] = types.StringNull()
		}
		if p.HttpGet.Port != nil {
			httpAttrs["port"] = types.Int64Value(*p.HttpGet.Port)
		} else {
			httpAttrs["port"] = types.Int64Null()
		}
		if p.HttpGet.Host != nil {
			httpAttrs["host"] = types.StringValue(*p.HttpGet.Host)
		} else {
			httpAttrs["host"] = types.StringNull()
		}
		if p.HttpGet.Scheme != nil {
			httpAttrs["scheme"] = types.StringValue(*p.HttpGet.Scheme)
		} else {
			httpAttrs["scheme"] = types.StringNull()
		}
		if len(p.HttpGet.HttpHeaders) > 0 {
			headerObjs := make([]attr.Value, len(p.HttpGet.HttpHeaders))
			for i, h := range p.HttpGet.HttpHeaders {
				obj, d := types.ObjectValue(httpHeaderAttrTypes, map[string]attr.Value{
					"name":  types.StringValue(h.Name),
					"value": types.StringValue(h.Value),
				})
				diags.Append(d...)
				headerObjs[i] = obj
			}
			headersList, d := types.ListValue(types.ObjectType{AttrTypes: httpHeaderAttrTypes}, headerObjs)
			diags.Append(d...)
			httpAttrs["http_headers"] = headersList
		} else {
			httpAttrs["http_headers"] = types.ListNull(types.ObjectType{AttrTypes: httpHeaderAttrTypes})
		}
		httpObj, d := types.ObjectValue(probeHttpGetAttrTypes, httpAttrs)
		diags.Append(d...)
		attrs["http_get"] = httpObj
	} else {
		attrs["http_get"] = types.ObjectNull(probeHttpGetAttrTypes)
	}

	// TCP Socket
	if p.TcpSocket != nil {
		tcpAttrs := map[string]attr.Value{
			"port": types.Int64Value(p.TcpSocket.Port),
		}
		if p.TcpSocket.Host != nil {
			tcpAttrs["host"] = types.StringValue(*p.TcpSocket.Host)
		} else {
			tcpAttrs["host"] = types.StringNull()
		}
		tcpObj, d := types.ObjectValue(probeTcpSocketAttrTypes, tcpAttrs)
		diags.Append(d...)
		attrs["tcp_socket"] = tcpObj
	} else {
		attrs["tcp_socket"] = types.ObjectNull(probeTcpSocketAttrTypes)
	}

	// Timing
	setOptionalInt64 := func(key string, val *int64) {
		if val != nil {
			attrs[key] = types.Int64Value(*val)
		} else {
			attrs[key] = types.Int64Null()
		}
	}
	setOptionalInt64("initial_delay_seconds", p.InitialDelaySeconds)
	setOptionalInt64("period_seconds", p.PeriodSeconds)
	setOptionalInt64("timeout_seconds", p.TimeoutSeconds)
	setOptionalInt64("success_threshold", p.SuccessThreshold)
	setOptionalInt64("failure_threshold", p.FailureThreshold)

	obj, d := types.ObjectValue(probeAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}

// HTTP methods.

func (r *applicationProcessResource) doCreateProcess(ctx context.Context, appID string, input *createProcessRequest) (*processResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling create process request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.client.BaseURL+"/applications/"+appID+"/processes", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(httpResp)
	}

	var process processResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&process); err != nil {
		return nil, fmt.Errorf("decoding process response: %w", err)
	}

	return &process, nil
}

func (r *applicationProcessResource) doGetProcess(ctx context.Context, appID, processID string) (*processResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL+"/applications/"+appID+"/processes/"+processID, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(httpResp)
	}

	var process processResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&process); err != nil {
		return nil, fmt.Errorf("decoding process response: %w", err)
	}

	return &process, nil
}

func (r *applicationProcessResource) doUpdateProcess(ctx context.Context, appID, processID string, input *updateProcessRequest) (*processResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling update process request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, r.client.BaseURL+"/applications/"+appID+"/processes/"+processID, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, parseHTTPError(httpResp)
	}

	var process processResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&process); err != nil {
		return nil, fmt.Errorf("decoding process response: %w", err)
	}

	return &process, nil
}

func (r *applicationProcessResource) doDeleteProcess(ctx context.Context, appID, processID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.client.BaseURL+"/applications/"+appID+"/processes/"+processID, nil)
	if err != nil {
		return err
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		return parseHTTPError(httpResp)
	}

	return nil
}

// parseHTTPError parses an error response and returns a structured APIError.
func parseHTTPError(resp *http.Response) error {
	var apiErr client.APIError
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		return &client.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code %d", resp.StatusCode),
		}
	}
	apiErr.StatusCode = resp.StatusCode
	return &apiErr
}
