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

type createPrivatePortRequest struct {
	ProcessID string `json:"process_id"`
	Port      int64  `json:"port"`
}

type privatePortResponse struct {
	ID        string  `json:"id"`
	ProcessID *string `json:"process_id"`
	Port      int64   `json:"port"`
	Status    *string `json:"status"`
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
}

func NewPrivatePortResource() resource.Resource {
	return &applicationPrivatePortResource{}
}

func (r *applicationPrivatePortResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_private_port"
}

func (r *applicationPrivatePortResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a private port for a Sevalla application.",
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
				Description: "The ID of the process this private port belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The port number.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the private port.",
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
	createReq := createPrivatePortRequest{
		ProcessID: plan.ProcessID.ValueString(),
		Port:      plan.Port.ValueInt64(),
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Private Port", err.Error())
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.client.BaseURL+"/applications/"+appID+"/private-ports", bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Private Port", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Private Port", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Creating Application Private Port", parseHTTPError(httpResp).Error())
		return
	}

	var port privatePortResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&port); err != nil {
		resp.Diagnostics.AddError("Error Creating Application Private Port", fmt.Sprintf("decoding response: %s", err))
		return
	}

	flattenPrivatePort(&port, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationPrivatePortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationPrivatePortResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL+"/applications/"+appID+"/private-ports/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application Private Port", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Application Private Port", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Error Reading Application Private Port", parseHTTPError(httpResp).Error())
		return
	}

	var port privatePortResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&port); err != nil {
		resp.Diagnostics.AddError("Error Reading Application Private Port", fmt.Sprintf("decoding response: %s", err))
		return
	}

	flattenPrivatePort(&port, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationPrivatePortResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Application private ports cannot be updated. Delete and recreate the resource instead.",
	)
}

func (r *applicationPrivatePortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationPrivatePortResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.client.BaseURL+"/applications/"+appID+"/private-ports/"+state.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Private Port", err.Error())
		return
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Private Port", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError("Error Deleting Application Private Port", parseHTTPError(httpResp).Error())
		return
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

func flattenPrivatePort(port *privatePortResponse, model *ApplicationPrivatePortResourceModel) {
	model.ID = types.StringValue(port.ID)
	if port.ProcessID != nil {
		model.ProcessID = types.StringValue(*port.ProcessID)
	}
	model.Port = types.Int64Value(port.Port)
	if port.Status != nil {
		model.Status = types.StringValue(*port.Status)
	} else {
		model.Status = types.StringNull()
	}
	if port.CreatedAt != nil {
		model.CreatedAt = types.StringValue(*port.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
	if port.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(*port.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}
}
