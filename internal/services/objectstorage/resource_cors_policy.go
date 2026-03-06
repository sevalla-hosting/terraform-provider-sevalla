package objectstorage

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &objectStorageCORSPolicyResource{}
	_ resource.ResourceWithConfigure   = &objectStorageCORSPolicyResource{}
	_ resource.ResourceWithImportState = &objectStorageCORSPolicyResource{}
)

type objectStorageCORSPolicyResource struct {
	client *client.SevallaClient
}

type ObjectStorageCORSPolicyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ObjectStorageID types.String `tfsdk:"object_storage_id"`
	AllowedOrigins  types.List   `tfsdk:"allowed_origins"`
	AllowedMethods  types.List   `tfsdk:"allowed_methods"`
	AllowedHeaders  types.List   `tfsdk:"allowed_headers"`
}

func NewCORSPolicyResource() resource.Resource {
	return &objectStorageCORSPolicyResource{}
}

func (r *objectStorageCORSPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_cors_policy"
}

func (r *objectStorageCORSPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CORS policy for a Sevalla object storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the CORS policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_storage_id": schema.StringAttribute{
				Description: "The ID of the object storage bucket.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allowed_origins": schema.ListAttribute{
				Description: "List of allowed origins.",
				Required:    true,
				ElementType: types.StringType,
			},
			"allowed_methods": schema.ListAttribute{
				Description: "List of allowed HTTP methods.",
				Required:    true,
				ElementType: types.StringType,
			},
			"allowed_headers": schema.ListAttribute{
				Description: "List of allowed HTTP headers.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *objectStorageCORSPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *objectStorageCORSPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectStorageCORSPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	osID := plan.ObjectStorageID.ValueString()
	createReq := &client.CreateCORSPolicyRequest{}

	resp.Diagnostics.Append(plan.AllowedOrigins.ElementsAs(ctx, &createReq.AllowedOrigins, false)...)
	resp.Diagnostics.Append(plan.AllowedMethods.ElementsAs(ctx, &createReq.AllowedMethods, false)...)
	if !plan.AllowedHeaders.IsNull() && !plan.AllowedHeaders.IsUnknown() {
		resp.Diagnostics.Append(plan.AllowedHeaders.ElementsAs(ctx, &createReq.AllowedHeaders, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CreateCORSPolicy(ctx, osID, createReq); err != nil {
		resp.Diagnostics.AddError("Error Creating Object Storage CORS Policy", err.Error())
		return
	}

	// POST response only returns {message}, so read back to get the created policy.
	policies, err := r.client.ListCORSPolicies(ctx, osID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Object Storage CORS Policy after create", err.Error())
		return
	}

	if len(policies) == 0 {
		resp.Diagnostics.AddError("Error Creating Object Storage CORS Policy", "No policies found after create")
		return
	}

	// Use the last policy in the list (most recently created).
	created := &policies[len(policies)-1]
	resp.Diagnostics.Append(flattenCORSPolicy(ctx, created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageCORSPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectStorageCORSPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	osID := state.ObjectStorageID.ValueString()
	policies, err := r.client.ListCORSPolicies(ctx, osID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Object Storage CORS Policy", err.Error())
		return
	}

	var found *client.CORSPolicy
	for i := range policies {
		if policies[i].ID == state.ID.ValueString() {
			found = &policies[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(flattenCORSPolicy(ctx, found, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageCORSPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ObjectStorageCORSPolicyResourceModel
	var state ObjectStorageCORSPolicyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	osID := state.ObjectStorageID.ValueString()
	updateReq := &client.UpdateCORSPolicyRequest{}

	resp.Diagnostics.Append(plan.AllowedOrigins.ElementsAs(ctx, &updateReq.AllowedOrigins, false)...)
	resp.Diagnostics.Append(plan.AllowedMethods.ElementsAs(ctx, &updateReq.AllowedMethods, false)...)
	if !plan.AllowedHeaders.IsNull() && !plan.AllowedHeaders.IsUnknown() {
		resp.Diagnostics.Append(plan.AllowedHeaders.ElementsAs(ctx, &updateReq.AllowedHeaders, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateCORSPolicy(ctx, osID, state.ID.ValueString(), updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Object Storage CORS Policy", err.Error())
		return
	}

	// PATCH response only returns {message}, so read back to get updated state.
	policies, err := r.client.ListCORSPolicies(ctx, osID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Object Storage CORS Policy after update", err.Error())
		return
	}

	var found *client.CORSPolicy
	for i := range policies {
		if policies[i].ID == state.ID.ValueString() {
			found = &policies[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Error Updating Object Storage CORS Policy", "Policy not found after update")
		return
	}

	resp.Diagnostics.Append(flattenCORSPolicy(ctx, found, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageCORSPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectStorageCORSPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCORSPolicy(ctx, state.ObjectStorageID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Object Storage CORS Policy", err.Error())
		return
	}
}

func (r *objectStorageCORSPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'object_storage_id/cors_policy_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("object_storage_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenCORSPolicy(ctx context.Context, policy *client.CORSPolicy, model *ObjectStorageCORSPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(policy.ID)

	origins, d := types.ListValueFrom(ctx, types.StringType, policy.AllowedOrigins)
	diags.Append(d...)
	model.AllowedOrigins = origins

	methods, d := types.ListValueFrom(ctx, types.StringType, policy.AllowedMethods)
	diags.Append(d...)
	model.AllowedMethods = methods

	if policy.AllowedHeaders != nil {
		headers, d := types.ListValueFrom(ctx, types.StringType, policy.AllowedHeaders)
		diags.Append(d...)
		model.AllowedHeaders = headers
	} else {
		model.AllowedHeaders = types.ListNull(types.StringType)
	}

	return diags
}
