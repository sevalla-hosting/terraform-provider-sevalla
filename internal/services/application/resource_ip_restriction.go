package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/listutil"
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
				Description: "The restriction type: `allow` lets only the listed IPs in, `deny` blocks them.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("allow", "deny")},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether IP restriction is enabled.",
				Required:    true,
			},
			"ip_list": schema.ListAttribute{
				Description: "IP addresses or CIDR ranges the restriction applies to (at most 500).",
				Required:    true,
				ElementType: types.StringType,
				Validators:  []validator.List{listvalidator.SizeAtMost(500)},
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

	result, err := r.client.UpdateApplicationIPRestriction(ctx, plan.ApplicationID.ValueString(), input)
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

	result, err := r.client.GetApplicationIPRestriction(ctx, state.ApplicationID.ValueString())
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

	result, err := r.client.UpdateApplicationIPRestriction(ctx, plan.ApplicationID.ValueString(), input)
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

	input := &client.IPRestriction{
		Type:      state.Type.ValueString(),
		IsEnabled: false,
		IPList:    []string{},
	}

	_, err := r.client.UpdateApplicationIPRestriction(ctx, state.ApplicationID.ValueString(), input)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Application IP Restriction", err.Error())
		return
	}
}

func (r *applicationIPRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("application_id"), req, resp)
}

func (r *applicationIPRestrictionResource) buildAppIPRestrictionInput(ctx context.Context, model *ApplicationIPRestrictionResourceModel) (*client.IPRestriction, diag.Diagnostics) {
	input := &client.IPRestriction{
		Type:      model.Type.ValueString(),
		IsEnabled: model.IsEnabled.ValueBool(),
		IPList:    []string{},
	}
	diags := model.IPList.ElementsAs(ctx, &input.IPList, false)
	return input, diags
}

func (r *applicationIPRestrictionResource) flattenAppIPRestriction(ctx context.Context, output *client.IPRestriction, model *ApplicationIPRestrictionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Type = types.StringValue(output.Type)
	model.IsEnabled = types.BoolValue(output.IsEnabled)

	var prior []string
	if !model.IPList.IsNull() && !model.IPList.IsUnknown() {
		diags.Append(model.IPList.ElementsAs(ctx, &prior, false)...)
	}
	ipList, d := types.ListValueFrom(ctx, types.StringType, listutil.KeepOrder(prior, output.IPList))
	diags.Append(d...)
	model.IPList = ipList

	return diags
}
