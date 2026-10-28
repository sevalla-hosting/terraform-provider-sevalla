package loadbalancer

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &loadBalancerDestinationResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerDestinationResource{}
	_ resource.ResourceWithImportState = &loadBalancerDestinationResource{}
)

type loadBalancerDestinationResource struct {
	client *client.SevallaClient
}

type LoadBalancerDestinationResourceModel struct {
	ID             types.String  `tfsdk:"id"`
	LoadBalancerID types.String  `tfsdk:"load_balancer_id"`
	ServiceType    types.String  `tfsdk:"service_type"`
	ServiceID      types.String  `tfsdk:"service_id"`
	Weight         types.Int64   `tfsdk:"weight"`
	URL            types.String  `tfsdk:"url"`
	Latitude       types.Float64 `tfsdk:"latitude"`
	Longitude      types.Float64 `tfsdk:"longitude"`
	IsEnabled      types.Bool    `tfsdk:"is_enabled"`
	CreatedAt      types.String  `tfsdk:"created_at"`
	UpdatedAt      types.String  `tfsdk:"updated_at"`
}

func NewDestinationResource() resource.Resource {
	return &loadBalancerDestinationResource{}
}

func (r *loadBalancerDestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_destination"
}

func (r *loadBalancerDestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a destination for a Sevalla load balancer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the destination.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"load_balancer_id": schema.StringAttribute{
				Description: "The ID of the load balancer.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_type": schema.StringAttribute{
				Description: "The type of the backing service. Valid values: `APP`, `STATIC_SITE`, `OBJECT_STORAGE`, `LOAD_BALANCER`, `EXTERNAL`.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("APP", "STATIC_SITE", "OBJECT_STORAGE", "EXTERNAL", "LOAD_BALANCER")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_id": schema.StringAttribute{
				Description: "The ID of the backing service. Required for every type except `EXTERNAL`, where it must not be set.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"weight": schema.Int64Attribute{
				Description: "The traffic weight for weighted routing (0-10000). Higher values receive proportionally more traffic.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Int64{int64validator.Between(0, 10000)},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The URL of an `EXTERNAL` destination.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"latitude": schema.Float64Attribute{
				Description: "The latitude for GEO routing (-90 to 90).",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Float64{float64validator.Between(-90, 90)},
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
					float64planmodifier.RequiresReplace(),
				},
			},
			"longitude": schema.Float64Attribute{
				Description: "The longitude for GEO routing (-180 to 180).",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Float64{float64validator.Between(-180, 180)},
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
					float64planmodifier.RequiresReplace(),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the destination receives traffic. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the destination was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the destination was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *loadBalancerDestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadBalancerDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbID := plan.LoadBalancerID.ValueString()
	createReq := &client.CreateLoadBalancerDestinationRequest{
		ServiceType: plan.ServiceType.ValueString(),
	}
	if !plan.ServiceID.IsNull() && !plan.ServiceID.IsUnknown() {
		createReq.ServiceID = plan.ServiceID.ValueString()
	}
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		v := int(plan.Weight.ValueInt64())
		createReq.Weight = &v
	}
	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		v := plan.URL.ValueString()
		createReq.URL = &v
	}
	if !plan.Latitude.IsNull() && !plan.Latitude.IsUnknown() {
		v := plan.Latitude.ValueFloat64()
		createReq.Latitude = &v
	}
	if !plan.Longitude.IsNull() && !plan.Longitude.IsUnknown() {
		v := plan.Longitude.ValueFloat64()
		createReq.Longitude = &v
	}
	dest, err := r.client.CreateLoadBalancerDestination(ctx, lbID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Load Balancer Destination", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("load_balancer_id"), lbID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), dest.ID)...)

	dest, err = r.setEnabled(ctx, lbID, dest, plan.IsEnabled)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Load Balancer Destination", err.Error())
		return
	}

	flattenLBDestination(dest, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *loadBalancerDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadBalancerDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.find(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Load Balancer Destination", err.Error())
		return
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenLBDestination(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is only reached when is_enabled changes; every other argument forces replacement.
func (r *loadBalancerDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadBalancerDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbID := state.LoadBalancerID.ValueString()
	dest, err := r.find(ctx, lbID, state.ID.ValueString())
	if err == nil && dest == nil {
		err = fmt.Errorf("destination %s not found", state.ID.ValueString())
	}
	if err == nil {
		dest, err = r.setEnabled(ctx, lbID, dest, plan.IsEnabled)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Load Balancer Destination", err.Error())
		return
	}

	flattenLBDestination(dest, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// find returns the destination with the given ID, or nil when there is none.
// The API has no endpoint to get a single destination.
func (r *loadBalancerDestinationResource) find(ctx context.Context, lbID, destID string) (*client.LoadBalancerDestination, error) {
	destinations, err := r.client.ListLoadBalancerDestinations(ctx, lbID)
	if err != nil {
		return nil, err
	}
	for i := range destinations {
		if destinations[i].ID == destID {
			return &destinations[i], nil
		}
	}
	return nil, nil
}

// setEnabled toggles the destination when its enabled state differs from
// want, which is skipped when want is unknown (not configured).
func (r *loadBalancerDestinationResource) setEnabled(ctx context.Context, lbID string, dest *client.LoadBalancerDestination, want types.Bool) (*client.LoadBalancerDestination, error) {
	if want.IsUnknown() || want.IsNull() || want.ValueBool() == dest.IsEnabled {
		return dest, nil
	}
	if err := r.client.ToggleLoadBalancerDestination(ctx, lbID, dest.ID); err != nil {
		return nil, fmt.Errorf("could not toggle destination %s: %w", dest.ID, err)
	}
	toggled, err := r.find(ctx, lbID, dest.ID)
	if err == nil && toggled == nil {
		err = fmt.Errorf("destination %s not found after toggling it", dest.ID)
	}
	return toggled, err
}

func (r *loadBalancerDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadBalancerDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadBalancerDestination(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Load Balancer Destination", err.Error())
		return
	}
}

func (r *loadBalancerDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'load_balancer_id/destination_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("load_balancer_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenLBDestination(dest *client.LoadBalancerDestination, model *LoadBalancerDestinationResourceModel) {
	model.ID = types.StringValue(dest.ID)
	if dest.ServiceID != nil {
		model.ServiceID = types.StringValue(*dest.ServiceID)
	} else {
		model.ServiceID = types.StringNull()
	}
	model.IsEnabled = types.BoolValue(dest.IsEnabled)

	if dest.ServiceType != nil {
		model.ServiceType = types.StringValue(*dest.ServiceType)
	} else {
		model.ServiceType = types.StringNull()
	}
	if dest.Weight != nil {
		model.Weight = types.Int64Value(int64(*dest.Weight))
	} else {
		model.Weight = types.Int64Null()
	}
	if dest.URL != nil {
		model.URL = types.StringValue(*dest.URL)
	} else {
		model.URL = types.StringNull()
	}
	if dest.Latitude != nil {
		model.Latitude = types.Float64Value(*dest.Latitude)
	} else {
		model.Latitude = types.Float64Null()
	}
	if dest.Longitude != nil {
		model.Longitude = types.Float64Value(*dest.Longitude)
	} else {
		model.Longitude = types.Float64Null()
	}
	if dest.CreatedAt != nil {
		model.CreatedAt = types.StringValue(*dest.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
	if dest.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(*dest.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}
}
