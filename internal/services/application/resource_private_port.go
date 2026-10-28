package application

import (
	"context"
	"fmt"
	"strings"

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

var (
	_ resource.Resource                = &applicationPrivatePortResource{}
	_ resource.ResourceWithConfigure   = &applicationPrivatePortResource{}
	_ resource.ResourceWithImportState = &applicationPrivatePortResource{}
)

type applicationPrivatePortResource struct {
	client *client.SevallaClient
}

type ApplicationPrivatePortResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	ProcessID     types.String `tfsdk:"process_id"`
	Port          types.Int64  `tfsdk:"port"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewPrivatePortResource() resource.Resource {
	return &applicationPrivatePortResource{}
}

func (r *applicationPrivatePortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_private_port"
}

func (r *applicationPrivatePortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a private port that exposes a port of a Sevalla application process inside the cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the private port.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this private port belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"process_id": schema.StringAttribute{
				Description: "The ID of the process to expose the port on.",
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
			"status": schema.StringAttribute{
				Description: "The status of the private port: `creating`, `ready`, `failed` or `deleting`.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the private port was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the private port was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *applicationPrivatePortResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationPrivatePortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationPrivatePortResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	port, err := r.client.CreatePrivatePort(ctx, appID, &client.CreatePortRequest{
		ProcessID: plan.ProcessID.ValueString(),
		Port:      plan.Port.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Private Port", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), appID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), port.ID)...)

	err = client.Poll(ctx, portReadyTimeout, func() (bool, error) {
		found, err := r.find(ctx, appID, port.ID)
		if err != nil || found == nil {
			return false, err
		}
		port = found
		switch port.Status {
		case "ready":
			return true, nil
		case "failed", "deleting":
			return false, fmt.Errorf("private port %s is %s", port.ID, port.Status)
		}
		return false, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for Private Port", err.Error())
		return
	}

	flattenPrivatePort(port, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationPrivatePortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationPrivatePortResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	port, err := r.find(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Private Port", err.Error())
		return
	}
	if port == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenPrivatePort(port, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// find returns the private port with the given ID, or nil when there is none.
// The API has no endpoint to get a single private port.
func (r *applicationPrivatePortResource) find(ctx context.Context, appID, portID string) (*client.PrivatePort, error) {
	ports, err := r.client.ListPrivatePorts(ctx, appID)
	if err != nil {
		return nil, err
	}
	for i := range ports {
		if ports[i].ID == portID {
			return &ports[i], nil
		}
	}
	return nil, nil
}

func (r *applicationPrivatePortResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Private ports cannot be updated. Change a required attribute to replace the resource.")
}

func (r *applicationPrivatePortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationPrivatePortResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePrivatePort(ctx, state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Private Port", err.Error())
	}
}

func (r *applicationPrivatePortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'application_id/private_port_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenPrivatePort(p *client.PrivatePort, model *ApplicationPrivatePortResourceModel) {
	model.ID = types.StringValue(p.ID)
	model.ProcessID = types.StringValue(p.ProcessID)
	model.Port = types.Int64Value(p.Port)
	model.Status = types.StringValue(p.Status)
	model.CreatedAt = types.StringValue(p.CreatedAt)
	model.UpdatedAt = types.StringValue(p.UpdatedAt)
}
