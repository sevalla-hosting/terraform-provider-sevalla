package application

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ID                    types.String `tfsdk:"id"`
	ApplicationID         types.String `tfsdk:"application_id"`
	DisplayName           types.String `tfsdk:"display_name"`
	Type                  types.String `tfsdk:"type"`
	Entrypoint            types.String `tfsdk:"entrypoint"`
	Port                  types.Int64  `tfsdk:"port"`
	ResourceTypeID        types.String `tfsdk:"resource_type_id"`
	ScalingStrategy       types.Object `tfsdk:"scaling_strategy"`
	Schedule              types.String `tfsdk:"schedule"`
	TimeZone              types.String `tfsdk:"time_zone"`
	ConcurrencyPolicy     types.String `tfsdk:"concurrency_policy"`
	ActiveDeadlineSeconds types.Int64  `tfsdk:"active_deadline_seconds"`
	JobStartPolicy        types.String `tfsdk:"job_start_policy"`
	IsIngressEnabled      types.Bool   `tfsdk:"is_ingress_enabled"`
	IngressProtocol       types.String `tfsdk:"ingress_protocol"`
	LivenessProbe         types.Object `tfsdk:"liveness_probe"`
	ReadinessProbe        types.Object `tfsdk:"readiness_probe"`
	Key                   types.String `tfsdk:"key"`
	ResourceTypeName      types.String `tfsdk:"resource_type_name"`
	CpuLimit              types.Int64  `tfsdk:"cpu_limit"`
	MemoryLimit           types.Int64  `tfsdk:"memory_limit"`
	InternalHostname      types.String `tfsdk:"internal_hostname"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
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

func NewProcessResource() resource.Resource {
	return &applicationProcessResource{}
}

func (r *applicationProcessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_process"
}

func (r *applicationProcessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	useState := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	resp.Schema = schema.Schema{
		Description: "Manages a process (web, worker, cron or job) of a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "The unique identifier of the process.",
				Computed:      true,
				PlanModifiers: useState,
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this process belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name for the process (1-255 characters). Cannot contain backticks, dollar signs, double quotes, backslashes or control characters.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"type": schema.StringAttribute{
				Description: "The process type. Valid values: `web`, `worker`, `cron`, `job`. Changing this creates a new process.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("web", "worker", "cron", "job")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entrypoint": schema.StringAttribute{
				Description:   "Command used to start the process (at most 1024 characters).",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{stringvalidator.LengthAtMost(1024)},
				PlanModifiers: useState,
			},
			"port": schema.Int64Attribute{
				Description: "Port number the process listens on (1-65535).",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 65535)},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The process resource type (machine size). Use the `sevalla_process_resource_types` data source to list them.",
				Required:    true,
			},
			"scaling_strategy": schema.SingleNestedAttribute{
				Description: "The scaling configuration for the process.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "The scaling type. Valid values: `manual`, `horizontal`.",
						Required:    true,
						Validators:  []validator.String{stringvalidator.OneOf("manual", "horizontal")},
					},
					"instance_count": schema.Int64Attribute{
						Description: "The number of instances (manual scaling only, 0-50).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(0, 50)},
					},
					"min_instance_count": schema.Int64Attribute{
						Description: "Minimum number of instances (horizontal scaling only, 0-50).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(0, 50)},
					},
					"max_instance_count": schema.Int64Attribute{
						Description: "Maximum number of instances (horizontal scaling only, 0-50).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(0, 50)},
					},
					"target_cpu_percent": schema.Int64Attribute{
						Description: "Target CPU utilization percentage for autoscaling (1-100).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 100)},
					},
					"target_memory_percent": schema.Int64Attribute{
						Description: "Target memory utilization percentage for autoscaling (1-100).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 100)},
					},
					"scale_up_interval_seconds": schema.Int64Attribute{
						Description: "Cooldown in seconds between scale-up events (1-60000).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 60000)},
					},
					"scale_up_increment": schema.Int64Attribute{
						Description: "Number of instances to add per scale-up event (1-5).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 5)},
					},
					"scale_down_interval_seconds": schema.Int64Attribute{
						Description: "Cooldown in seconds between scale-down events (1-60000).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 60000)},
					},
					"scale_down_increment": schema.Int64Attribute{
						Description: "Number of instances to remove per scale-down event (1-5).",
						Optional:    true,
						Validators:  []validator.Int64{int64validator.Between(1, 5)},
					},
				},
			},
			"schedule": schema.StringAttribute{
				Description: "Cron expression for the schedule (required for cron processes, at most 100 characters).",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtMost(100)},
			},
			"time_zone": schema.StringAttribute{
				Description: "IANA time zone for the cron schedule (e.g. `America/New_York`).",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtMost(100)},
			},
			"concurrency_policy": schema.StringAttribute{
				Description:   "What a cron process does when the previous run is still active: `Forbid` skips the new run, `Replace` cancels the running one.",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{stringvalidator.OneOf("Forbid", "Replace")},
				PlanModifiers: useState,
			},
			"active_deadline_seconds": schema.Int64Attribute{
				Description: "Maximum run time in seconds of a cron run or job. When unset or `0`, cron runs have no limit and jobs use the default of 600 seconds.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(0)},
			},
			"job_start_policy": schema.StringAttribute{
				Description: "When a job process runs relative to deployments. Valid values: `beforeDeployment`, `afterSuccessDeployment`, `afterFailedDeployment`.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.OneOf("beforeDeployment", "afterSuccessDeployment", "afterFailedDeployment")},
			},
			"is_ingress_enabled": schema.BoolAttribute{
				Description: "Whether external traffic is routed to the process (web processes only).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ingress_protocol": schema.StringAttribute{
				Description:   "The ingress protocol (web processes only). Valid values: `http`, `grpc`.",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{stringvalidator.OneOf("http", "grpc")},
				PlanModifiers: useState,
			},
			"liveness_probe": schema.SingleNestedAttribute{
				Description: "Liveness probe (web processes only). Only HTTP GET probes can be configured.",
				Optional:    true,
				Computed:    true,
				Attributes:  probeSchemaAttributes(),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"readiness_probe": schema.SingleNestedAttribute{
				Description: "Readiness probe (web processes only). Only HTTP GET probes can be configured.",
				Optional:    true,
				Computed:    true,
				Attributes:  probeSchemaAttributes(),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description:   "Unique process key within the application.",
				Computed:      true,
				PlanModifiers: useState,
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
				Description:   "Internal Kubernetes DNS hostname for service-to-service communication.",
				Computed:      true,
				PlanModifiers: useState,
			},
			"created_at": schema.StringAttribute{
				Description:   "The timestamp when the process was created.",
				Computed:      true,
				PlanModifiers: useState,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the process was last updated.",
				Computed:    true,
			},
		},
	}
}

func probeSchemaAttributes() map[string]schema.Attribute {
	optionalInt := func(description string, min int64) schema.Int64Attribute {
		return schema.Int64Attribute{
			Description: description,
			Optional:    true,
			Computed:    true,
			Validators:  []validator.Int64{int64validator.AtLeast(min)},
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		}
	}
	return map[string]schema.Attribute{
		"exec": schema.SingleNestedAttribute{
			Description: "Exec probe that runs a command inside the container. Read-only: the API only accepts HTTP GET probes.",
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"command": schema.ListAttribute{
					Description: "Command to execute.",
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
		"http_get": schema.SingleNestedAttribute{
			Description: "HTTP GET probe.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"path": schema.StringAttribute{
					Description: "Path to probe. A probe without a path is not saved.",
					Required:    true,
				},
				"port": schema.Int64Attribute{
					Description: "Port to probe (1-65535).",
					Optional:    true,
					Computed:    true,
					Validators:  []validator.Int64{int64validator.Between(1, 65535)},
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"host": schema.StringAttribute{
					Description: "Hostname for the probe request.",
					Optional:    true,
					Computed:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"scheme": schema.StringAttribute{
					Description: "URI scheme. Valid values: `HTTP`, `HTTPS`.",
					Optional:    true,
					Computed:    true,
					Validators:  []validator.String{stringvalidator.OneOf("HTTP", "HTTPS")},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
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
			Description: "TCP socket probe. Read-only: the API only accepts HTTP GET probes.",
			Computed:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
			Attributes: map[string]schema.Attribute{
				"host": schema.StringAttribute{
					Description: "Hostname to connect to.",
					Computed:    true,
				},
				"port": schema.Int64Attribute{
					Description: "Port to connect to.",
					Computed:    true,
				},
			},
		},
		"initial_delay_seconds": optionalInt("Seconds after the container starts before the probe starts.", 0),
		"period_seconds":        optionalInt("How often (in seconds) to perform the probe.", 1),
		"timeout_seconds":       optionalInt("Seconds after which the probe times out. The API defaults it to 120.", 1),
		"success_threshold":     optionalInt("Consecutive successes required to be considered healthy.", 1),
		"failure_threshold":     optionalInt("Consecutive failures required to be considered unhealthy.", 1),
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

	appID := plan.ApplicationID.ValueString()
	process, err := r.client.CreateProcess(ctx, appID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Process", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), appID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), process.ID)...)

	// The create endpoint does not accept the ingress settings; apply them with an update.
	updateReq := &client.UpdateProcessRequest{}
	if v := plan.IsIngressEnabled; !v.IsUnknown() && !v.IsNull() && v.ValueBool() != process.IsIngressEnabled {
		updateReq.IsIngressEnabled = v.ValueBoolPointer()
	}
	if v := plan.IngressProtocol; !v.IsUnknown() && !v.IsNull() && (process.IngressProtocol == nil || v.ValueString() != *process.IngressProtocol) {
		updateReq.IngressProtocol = v.ValueStringPointer()
	}
	if updateReq.IsIngressEnabled != nil || updateReq.IngressProtocol != nil {
		process, err = r.client.UpdateProcess(ctx, appID, process.ID, updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Application Process After Create", err.Error())
			return
		}
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

	process, err := r.client.GetProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
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

	process, err := r.client.UpdateProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString(), updateReq)
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

	err := r.client.DeleteProcess(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
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

func knownString(v types.String) *string {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}

func knownInt64(v types.Int64) *int64 {
	if v.IsUnknown() {
		return nil
	}
	return v.ValueInt64Pointer()
}

func buildCreateProcessRequest(ctx context.Context, plan *ApplicationProcessResourceModel) (*client.CreateProcessRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &client.CreateProcessRequest{
		DisplayName:           plan.DisplayName.ValueString(),
		Type:                  plan.Type.ValueString(),
		ResourceTypeID:        plan.ResourceTypeID.ValueString(),
		Entrypoint:            knownString(plan.Entrypoint),
		Port:                  knownInt64(plan.Port),
		Schedule:              knownString(plan.Schedule),
		TimeZone:              knownString(plan.TimeZone),
		ConcurrencyPolicy:     knownString(plan.ConcurrencyPolicy),
		ActiveDeadlineSeconds: knownInt64(plan.ActiveDeadlineSeconds),
		JobStartPolicy:        knownString(plan.JobStartPolicy),
	}

	ss, d := buildScalingStrategy(plan.ScalingStrategy)
	diags.Append(d...)
	req.ScalingStrategy = ss

	p, d := buildProbe(ctx, plan.LivenessProbe)
	diags.Append(d...)
	req.LivenessProbe = p

	p, d = buildProbe(ctx, plan.ReadinessProbe)
	diags.Append(d...)
	req.ReadinessProbe = p

	return req, diags
}

// buildUpdateProcessRequest sends only the attributes whose planned value is
// known and differs from state.
func buildUpdateProcessRequest(ctx context.Context, plan, state *ApplicationProcessResourceModel) (*client.UpdateProcessRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.UpdateProcessRequest{}

	changed := func(p, s attr.Value) bool { return !p.IsUnknown() && !p.Equal(s) }
	empty := ""
	var zero int64

	if changed(plan.DisplayName, state.DisplayName) {
		req.DisplayName = plan.DisplayName.ValueStringPointer()
	}
	if changed(plan.Entrypoint, state.Entrypoint) {
		req.Entrypoint = plan.Entrypoint.ValueStringPointer()
	}
	if changed(plan.Port, state.Port) {
		req.Port = plan.Port.ValueInt64Pointer()
	}
	if changed(plan.ResourceTypeID, state.ResourceTypeID) {
		req.ResourceTypeID = plan.ResourceTypeID.ValueStringPointer()
	}
	// The API clears these when sent an empty value.
	if changed(plan.Schedule, state.Schedule) {
		req.Schedule = &empty
		if !plan.Schedule.IsNull() {
			req.Schedule = plan.Schedule.ValueStringPointer()
		}
	}
	if changed(plan.TimeZone, state.TimeZone) {
		req.TimeZone = &empty
		if !plan.TimeZone.IsNull() {
			req.TimeZone = plan.TimeZone.ValueStringPointer()
		}
	}
	if changed(plan.ActiveDeadlineSeconds, state.ActiveDeadlineSeconds) {
		req.ActiveDeadlineSeconds = &zero
		if !plan.ActiveDeadlineSeconds.IsNull() {
			req.ActiveDeadlineSeconds = plan.ActiveDeadlineSeconds.ValueInt64Pointer()
		}
	}
	if changed(plan.ConcurrencyPolicy, state.ConcurrencyPolicy) {
		req.ConcurrencyPolicy = plan.ConcurrencyPolicy.ValueStringPointer()
	}
	if changed(plan.JobStartPolicy, state.JobStartPolicy) {
		req.JobStartPolicy = plan.JobStartPolicy.ValueStringPointer()
	}
	if changed(plan.IsIngressEnabled, state.IsIngressEnabled) {
		req.IsIngressEnabled = plan.IsIngressEnabled.ValueBoolPointer()
	}
	if changed(plan.IngressProtocol, state.IngressProtocol) {
		req.IngressProtocol = plan.IngressProtocol.ValueStringPointer()
	}
	if changed(plan.ScalingStrategy, state.ScalingStrategy) {
		ss, d := buildScalingStrategy(plan.ScalingStrategy)
		diags.Append(d...)
		req.ScalingStrategy = ss
	}
	if changed(plan.LivenessProbe, state.LivenessProbe) {
		p, d := buildProbe(ctx, plan.LivenessProbe)
		diags.Append(d...)
		req.LivenessProbe = p
	}
	if changed(plan.ReadinessProbe, state.ReadinessProbe) {
		p, d := buildProbe(ctx, plan.ReadinessProbe)
		diags.Append(d...)
		req.ReadinessProbe = p
	}

	return req, diags
}

func int64Attr(attrs map[string]attr.Value, key string) *int64 {
	v, ok := attrs[key].(types.Int64)
	if !ok || v.IsUnknown() {
		return nil
	}
	return v.ValueInt64Pointer()
}

func buildScalingStrategy(obj types.Object) (*client.ScalingStrategy, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	attrs := obj.Attributes()
	ss := &client.ScalingStrategy{Type: attrs["type"].(types.String).ValueString()}

	switch ss.Type {
	case "manual":
		ss.Config.InstanceCount = int64Attr(attrs, "instance_count")
		if ss.Config.InstanceCount == nil {
			diags.AddAttributeError(path.Root("scaling_strategy").AtName("instance_count"), "Invalid Scaling Strategy", "instance_count is required for manual scaling.")
		}
	case "horizontal":
		ss.Config = client.ScalingConfig{
			MinInstanceCount:         int64Attr(attrs, "min_instance_count"),
			MaxInstanceCount:         int64Attr(attrs, "max_instance_count"),
			TargetCPUPercent:         int64Attr(attrs, "target_cpu_percent"),
			TargetMemoryPercent:      int64Attr(attrs, "target_memory_percent"),
			ScaleUpIntervalSeconds:   int64Attr(attrs, "scale_up_interval_seconds"),
			ScaleUpIncrement:         int64Attr(attrs, "scale_up_increment"),
			ScaleDownIntervalSeconds: int64Attr(attrs, "scale_down_interval_seconds"),
			ScaleDownIncrement:       int64Attr(attrs, "scale_down_increment"),
		}
		if ss.Config.MinInstanceCount == nil || ss.Config.MaxInstanceCount == nil {
			diags.AddAttributeError(path.Root("scaling_strategy"), "Invalid Scaling Strategy", "min_instance_count and max_instance_count are required for horizontal scaling.")
		}
	}

	return ss, diags
}

// buildProbe converts a configured probe to an API probe. Only HTTP GET probes
// can be written, so exec and tcp_socket are never sent.
func buildProbe(ctx context.Context, obj types.Object) (*client.Probe, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}

	attrs := obj.Attributes()
	probe := &client.Probe{
		InitialDelaySeconds: int64Attr(attrs, "initial_delay_seconds"),
		PeriodSeconds:       int64Attr(attrs, "period_seconds"),
		TimeoutSeconds:      int64Attr(attrs, "timeout_seconds"),
		SuccessThreshold:    int64Attr(attrs, "success_threshold"),
		FailureThreshold:    int64Attr(attrs, "failure_threshold"),
	}

	if httpObj, ok := attrs["http_get"].(types.Object); ok && !httpObj.IsNull() && !httpObj.IsUnknown() {
		httpAttrs := httpObj.Attributes()
		probe.HTTPGet = &client.ProbeHTTPGet{
			Path:   knownString(httpAttrs["path"].(types.String)),
			Port:   int64Attr(httpAttrs, "port"),
			Host:   knownString(httpAttrs["host"].(types.String)),
			Scheme: knownString(httpAttrs["scheme"].(types.String)),
		}
		if headers, ok := httpAttrs["http_headers"].(types.List); ok && !headers.IsNull() && !headers.IsUnknown() {
			var headerObjs []types.Object
			diags.Append(headers.ElementsAs(ctx, &headerObjs, false)...)
			for _, h := range headerObjs {
				ha := h.Attributes()
				probe.HTTPGet.HTTPHeaders = append(probe.HTTPGet.HTTPHeaders, client.HTTPHeader{
					Name:  ha["name"].(types.String).ValueString(),
					Value: ha["value"].(types.String).ValueString(),
				})
			}
		}
	}

	return probe, diags
}

// Flatten helpers.

func optionalInt64(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

// nonEmptyString maps nil and "" to null.
func nonEmptyString(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func flattenProcess(ctx context.Context, p *client.Process, model *ApplicationProcessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(p.ID)
	model.Key = types.StringValue(p.Key)
	if p.Type != nil {
		model.Type = types.StringValue(*p.Type)
	}
	if p.DisplayName != nil {
		model.DisplayName = types.StringValue(*p.DisplayName)
	}
	if p.ResourceTypeID != nil {
		model.ResourceTypeID = types.StringValue(*p.ResourceTypeID)
	}
	model.Entrypoint = types.StringValue(p.Entrypoint)
	model.Port = optionalInt64(p.Port)
	model.IsIngressEnabled = types.BoolValue(p.IsIngressEnabled)
	model.IngressProtocol = optionalString(p.IngressProtocol)
	model.ResourceTypeName = optionalString(p.ResourceTypeName)
	model.CpuLimit = optionalInt64(p.CPULimit)
	model.MemoryLimit = optionalInt64(p.MemoryLimit)
	model.Schedule = nonEmptyString(p.Schedule)
	model.TimeZone = nonEmptyString(p.TimeZone)
	model.ConcurrencyPolicy = optionalString(p.ConcurrencyPolicy)
	model.JobStartPolicy = nonEmptyString(p.JobStartPolicy)
	model.InternalHostname = optionalString(p.InternalHostname)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)

	// The API reports an unset limit as null; keep a configured 0, which means the same.
	if p.ActiveDeadlineSeconds != nil || model.ActiveDeadlineSeconds.IsUnknown() || model.ActiveDeadlineSeconds.ValueInt64() != 0 {
		model.ActiveDeadlineSeconds = optionalInt64(p.ActiveDeadlineSeconds)
	}

	if p.ScalingStrategy != nil {
		cfg := p.ScalingStrategy.Config
		obj, d := types.ObjectValue(scalingStrategyAttrTypes, map[string]attr.Value{
			"type":                        types.StringValue(p.ScalingStrategy.Type),
			"instance_count":              optionalInt64(cfg.InstanceCount),
			"min_instance_count":          optionalInt64(cfg.MinInstanceCount),
			"max_instance_count":          optionalInt64(cfg.MaxInstanceCount),
			"target_cpu_percent":          optionalInt64(cfg.TargetCPUPercent),
			"target_memory_percent":       optionalInt64(cfg.TargetMemoryPercent),
			"scale_up_interval_seconds":   optionalInt64(cfg.ScaleUpIntervalSeconds),
			"scale_up_increment":          optionalInt64(cfg.ScaleUpIncrement),
			"scale_down_interval_seconds": optionalInt64(cfg.ScaleDownIntervalSeconds),
			"scale_down_increment":        optionalInt64(cfg.ScaleDownIncrement),
		})
		diags.Append(d...)
		model.ScalingStrategy = obj
	} else {
		model.ScalingStrategy = types.ObjectNull(scalingStrategyAttrTypes)
	}

	lp, d := flattenProbe(ctx, p.LivenessProbe)
	diags.Append(d...)
	model.LivenessProbe = lp

	rp, d := flattenProbe(ctx, p.ReadinessProbe)
	diags.Append(d...)
	model.ReadinessProbe = rp

	return diags
}

func flattenProbe(ctx context.Context, p *client.Probe) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The API returns {} for a process without the probe.
	if p == nil || reflect.ValueOf(*p).IsZero() {
		return types.ObjectNull(probeAttrTypes), diags
	}

	attrs := map[string]attr.Value{
		"exec":                  types.ObjectNull(probeExecAttrTypes),
		"http_get":              types.ObjectNull(probeHttpGetAttrTypes),
		"tcp_socket":            types.ObjectNull(probeTcpSocketAttrTypes),
		"initial_delay_seconds": optionalInt64(p.InitialDelaySeconds),
		"period_seconds":        optionalInt64(p.PeriodSeconds),
		"timeout_seconds":       optionalInt64(p.TimeoutSeconds),
		"success_threshold":     optionalInt64(p.SuccessThreshold),
		"failure_threshold":     optionalInt64(p.FailureThreshold),
	}

	if p.Exec != nil {
		cmdList, d := types.ListValueFrom(ctx, types.StringType, p.Exec.Command)
		diags.Append(d...)
		execObj, d := types.ObjectValue(probeExecAttrTypes, map[string]attr.Value{"command": cmdList})
		diags.Append(d...)
		attrs["exec"] = execObj
	}

	if p.HTTPGet != nil {
		headerType := types.ObjectType{AttrTypes: httpHeaderAttrTypes}
		headers := types.ListNull(headerType)
		if len(p.HTTPGet.HTTPHeaders) > 0 {
			headerObjs := make([]attr.Value, len(p.HTTPGet.HTTPHeaders))
			for i, h := range p.HTTPGet.HTTPHeaders {
				obj, d := types.ObjectValue(httpHeaderAttrTypes, map[string]attr.Value{
					"name":  types.StringValue(h.Name),
					"value": types.StringValue(h.Value),
				})
				diags.Append(d...)
				headerObjs[i] = obj
			}
			list, d := types.ListValue(headerType, headerObjs)
			diags.Append(d...)
			headers = list
		}
		httpObj, d := types.ObjectValue(probeHttpGetAttrTypes, map[string]attr.Value{
			"path":         optionalString(p.HTTPGet.Path),
			"port":         optionalInt64(p.HTTPGet.Port),
			"host":         optionalString(p.HTTPGet.Host),
			"scheme":       optionalString(p.HTTPGet.Scheme),
			"http_headers": headers,
		})
		diags.Append(d...)
		attrs["http_get"] = httpObj
	}

	if p.TCPSocket != nil {
		tcpObj, d := types.ObjectValue(probeTcpSocketAttrTypes, map[string]attr.Value{
			"host": optionalString(p.TCPSocket.Host),
			"port": types.Int64Value(p.TCPSocket.Port),
		})
		diags.Append(d...)
		attrs["tcp_socket"] = tcpObj
	}

	obj, d := types.ObjectValue(probeAttrTypes, attrs)
	diags.Append(d...)
	return obj, diags
}
