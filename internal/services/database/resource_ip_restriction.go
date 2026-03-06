package database

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &databaseIPRestrictionResource{}
	_ resource.ResourceWithConfigure   = &databaseIPRestrictionResource{}
	_ resource.ResourceWithImportState = &databaseIPRestrictionResource{}
)

type databaseIPRestrictionResource struct {
	client *client.SevallaClient
}

type DatabaseIPRestrictionResourceModel struct {
	DatabaseID types.String `tfsdk:"database_id"`
	Type       types.String `tfsdk:"type"`
	IsEnabled  types.Bool   `tfsdk:"is_enabled"`
	IPList     types.List   `tfsdk:"ip_list"`
}

func NewIPRestrictionResource() resource.Resource {
	return &databaseIPRestrictionResource{}
}

func (r *databaseIPRestrictionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_ip_restriction"
}

func (r *databaseIPRestrictionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IP restrictions for a Sevalla database.",
		Attributes: map[string]schema.Attribute{
			"database_id": schema.StringAttribute{
				Description: "The ID of the database. Acts as the resource identifier.",
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

func (r *databaseIPRestrictionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseIPRestrictionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseIPRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := buildDBIPRestrictionInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateDatabaseIPRestriction(ctx, plan.DatabaseID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Database IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDBIPRestriction(ctx, result, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseIPRestrictionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseIPRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetDatabaseIPRestriction(ctx, state.DatabaseID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Database IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDBIPRestriction(ctx, result, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseIPRestrictionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DatabaseIPRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := buildDBIPRestrictionInput(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateDatabaseIPRestriction(ctx, plan.DatabaseID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Database IP Restriction", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDBIPRestriction(ctx, result, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseIPRestrictionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseIPRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &client.IPRestriction{
		Type:      state.Type.ValueString(),
		IsEnabled: false,
		IPList:    []string{},
	}

	_, err := r.client.UpdateDatabaseIPRestriction(ctx, state.DatabaseID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Database IP Restriction", err.Error())
		return
	}
}

func (r *databaseIPRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("database_id"), req, resp)
}

func buildDBIPRestrictionInput(ctx context.Context, model *DatabaseIPRestrictionResourceModel) (*client.IPRestriction, diag.Diagnostics) {
	var diags diag.Diagnostics

	input := &client.IPRestriction{
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

func flattenDBIPRestriction(ctx context.Context, output *client.IPRestriction, model *DatabaseIPRestrictionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.Type = types.StringValue(output.Type)
	model.IsEnabled = types.BoolValue(output.IsEnabled)

	ipList, d := types.ListValueFrom(ctx, types.StringType, output.IPList)
	diags.Append(d...)
	model.IPList = ipList

	return diags
}
