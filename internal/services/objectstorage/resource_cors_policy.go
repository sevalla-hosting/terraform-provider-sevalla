package objectstorage

import (
	"context"
	"fmt"
	"strings"

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
				Description: "Origins allowed by the rule (1-100). Use `*` to allow all origins.",
				Required:    true,
				ElementType: types.StringType,
				Validators:  []validator.List{listvalidator.SizeBetween(1, 100)},
			},
			"allowed_methods": schema.ListAttribute{
				Description: "HTTP methods allowed by the rule. Valid values: `GET`, `PUT`, `POST`, `DELETE`, `HEAD`.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("GET", "PUT", "POST", "DELETE", "HEAD")),
				},
			},
			"allowed_headers": schema.ListAttribute{
				Description: "HTTP headers allowed in CORS requests (1-100). Omit to allow no specific headers.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.List{listvalidator.SizeBetween(1, 100)},
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
	createReq, diags := buildCORSPolicyRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The create response has no ID, so find the new policy by comparing lists.
	before, err := r.client.ListCORSPolicies(ctx, osID)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Object Storage CORS Policy", err.Error())
		return
	}
	known := make(map[string]bool, len(before))
	for _, p := range before {
		known[p.ID] = true
	}

	if err := r.client.CreateCORSPolicy(ctx, osID, createReq); err != nil {
		resp.Diagnostics.AddError("Error Creating Object Storage CORS Policy", err.Error())
		return
	}

	after, err := r.client.ListCORSPolicies(ctx, osID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Object Storage CORS Policy after create", err.Error())
		return
	}
	var created []client.CORSPolicy
	for _, p := range after {
		if !known[p.ID] {
			created = append(created, p)
		}
	}
	// Policies created concurrently on the same bucket are told apart by their rules.
	if len(created) > 1 {
		var matching []client.CORSPolicy
		for _, p := range created {
			if listutil.SameElements(p.AllowedOrigins, createReq.AllowedOrigins) && listutil.SameElements(p.AllowedMethods, createReq.AllowedMethods) && listutil.SameElements(p.AllowedHeaders, createReq.AllowedHeaders) {
				matching = append(matching, p)
			}
		}
		created = matching
	}
	if len(created) != 1 {
		resp.Diagnostics.AddError("Error Creating Object Storage CORS Policy", fmt.Sprintf("Could not identify the created policy: %d candidates found.", len(created)))
		return
	}

	resp.Diagnostics.Append(flattenCORSPolicy(ctx, &created[0], &plan)...)
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
	updateReq, diags := buildCORSPolicyRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Send an empty list to remove headers that were set before.
	if updateReq.AllowedHeaders == nil {
		updateReq.AllowedHeaders = []string{}
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
	if err != nil && !client.IsNotFound(err) {
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

func buildCORSPolicyRequest(ctx context.Context, plan *ObjectStorageCORSPolicyResourceModel) (*client.CORSPolicyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := &client.CORSPolicyRequest{}
	diags.Append(plan.AllowedOrigins.ElementsAs(ctx, &req.AllowedOrigins, false)...)
	diags.Append(plan.AllowedMethods.ElementsAs(ctx, &req.AllowedMethods, false)...)
	if !plan.AllowedHeaders.IsNull() && !plan.AllowedHeaders.IsUnknown() {
		diags.Append(plan.AllowedHeaders.ElementsAs(ctx, &req.AllowedHeaders, false)...)
	}
	return req, diags
}

func flattenCORSPolicy(ctx context.Context, policy *client.CORSPolicy, model *ObjectStorageCORSPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue(policy.ID)
	model.AllowedOrigins = flattenStringList(ctx, model.AllowedOrigins, policy.AllowedOrigins, &diags)
	model.AllowedMethods = flattenStringList(ctx, model.AllowedMethods, policy.AllowedMethods, &diags)
	model.AllowedHeaders = types.ListNull(types.StringType)
	if len(policy.AllowedHeaders) > 0 {
		model.AllowedHeaders = flattenStringList(ctx, model.AllowedHeaders, policy.AllowedHeaders, &diags)
	}

	return diags
}

// flattenStringList keeps the order of prior when it holds the same elements
// as got, so a server-side reordering does not show up as a diff.
func flattenStringList(ctx context.Context, prior types.List, got []string, diags *diag.Diagnostics) types.List {
	var priorValues []string
	if !prior.IsNull() && !prior.IsUnknown() {
		diags.Append(prior.ElementsAs(ctx, &priorValues, false)...)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, listutil.KeepOrder(priorValues, got))
	diags.Append(d...)
	return list
}
