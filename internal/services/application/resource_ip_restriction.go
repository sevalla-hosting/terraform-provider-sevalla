package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	_ resource.Resource                = &applicationIPRestrictionResource{}
	_ resource.ResourceWithConfigure   = &applicationIPRestrictionResource{}
	_ resource.ResourceWithImportState = &applicationIPRestrictionResource{}
)

type applicationIPRestrictionResource struct {
	client *client.SevallaClient
}

type ApplicationIPRestrictionResourceModel struct {
	ApplicationID types.String `tfsdk:"application_id"`
	Type          types.String `tfsdk:"type"`
	IsEnabled     types.Bool   `tfsdk:"is_enabled"`
	IPList        types.List   `tfsdk:"ip_list"`
}

func NewIPRestrictionResource() resource.Resource {
	return &applicationIPRestrictionResource{}
}

func (r *applicationIPRestrictionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_ip_restriction"
}

func (r *applicationIPRestrictionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IP restrictions for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "The ID of the application. Acts as the resource identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The restriction type. Valid values: allow, deny.",
				Required:    true,
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether IP restriction is enabled.",
				Required:    true,
			},
			"ip_list": schema.ListAttribute{
				Description: "List of IP addresses or CIDR ranges.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *applicationIPRestrictionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationIPRestrictionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationIPRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := r.buildAppIPRestrictionInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.doUpdateAppIPRestriction(ctx, plan.ApplicationID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(r.flattenAppIPRestriction(ctx, result, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationIPRestrictionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationIPRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.doGetAppIPRestriction(ctx, state.ApplicationID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(r.flattenAppIPRestriction(ctx, result, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationIPRestrictionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationIPRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := r.buildAppIPRestrictionInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.doUpdateAppIPRestriction(ctx, plan.ApplicationID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(r.flattenAppIPRestriction(ctx, result, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationIPRestrictionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationIPRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &appIPRestrictionInput{
		Type:      state.Type.ValueString(),
		IsEnabled: false,
		IPList:    []string{},
	}

	_, err := r.doUpdateAppIPRestriction(ctx, state.ApplicationID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application IP Restriction", err.Error())
		return
	}
}

func (r *applicationIPRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}

type appIPRestrictionInput struct {
	Type      string   `json:"type"`
	IsEnabled bool     `json:"is_enabled"`
	IPList    []string `json:"ip_list"`
}

type appIPRestrictionOutput struct {
	Type      string   `json:"type"`
	IsEnabled bool     `json:"is_enabled"`
	IPList    []string `json:"ip_list"`
}

func (r *applicationIPRestrictionResource) buildAppIPRestrictionInput(ctx context.Context, model *ApplicationIPRestrictionResourceModel) (*appIPRestrictionInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &appIPRestrictionInput{
		Type:      model.Type.ValueString(),
		IsEnabled: model.IsEnabled.ValueBool(),
	}

	diags.Append(model.IPList.ElementsAs(ctx, &input.IPList, false)...)
	if diags.HasError() {
		return nil, diags
	}

	if input.IPList == nil {
		input.IPList = []string{}
	}

	return input, diags
}

func (r *applicationIPRestrictionResource) flattenAppIPRestriction(ctx context.Context, output *appIPRestrictionOutput, model *ApplicationIPRestrictionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Type = types.StringValue(output.Type)
	model.IsEnabled = types.BoolValue(output.IsEnabled)

	ipList, d := types.ListValueFrom(ctx, types.StringType, output.IPList)
	diags.Append(d...)
	model.IPList = ipList

	return diags
}

func (r *applicationIPRestrictionResource) doGetAppIPRestriction(ctx context.Context, appID string) (*appIPRestrictionOutput, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, r.client.BaseURL+"/applications/"+appID+"/ip-restriction", nil)
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

	var output appIPRestrictionOutput
	if err := json.NewDecoder(httpResp.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decoding ip restriction response: %w", err)
	}

	return &output, nil
}

func (r *applicationIPRestrictionResource) doUpdateAppIPRestriction(ctx context.Context, appID string, input *appIPRestrictionInput) (*appIPRestrictionOutput, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshaling ip restriction request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, r.client.BaseURL+"/applications/"+appID+"/ip-restriction", bytes.NewReader(body))
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

	var output appIPRestrictionOutput
	if err := json.NewDecoder(httpResp.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decoding ip restriction response: %w", err)
	}

	return &output, nil
}
