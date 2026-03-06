package objectstorage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &objectStorageResource{}
	_ resource.ResourceWithConfigure   = &objectStorageResource{}
	_ resource.ResourceWithImportState = &objectStorageResource{}
)

type objectStorageResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &objectStorageResource{}
}

func (r *objectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage"
}

func (r *objectStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla object storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the object storage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the object storage.",
				Required:    true,
			},
			"location": schema.StringAttribute{
				Description: "Geographic hint for where most data access occurs. Valid values: apac, eeur, enam, oc, weur, wnam. Defaults to enam.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"jurisdiction": schema.StringAttribute{
				Description: "Data residency jurisdiction. Valid values: default, eu, fedramp. Defaults to default.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_access": schema.BoolAttribute{
				Description: "Whether to enable a public CDN domain for this bucket. Defaults to false.",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project to group this bucket under.",
				Optional:    true,
			},
			// Computed-only attributes
			"name": schema.StringAttribute{
				Description: "The system-generated name of the object storage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The public CDN domain for accessing objects. Null if public access is not enabled.",
				Computed:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The S3-compatible API endpoint URL for the object storage.",
				Computed:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "The access key for S3-compatible API authentication. Only returned on create.",
				Computed:    true,
				Sensitive:   true,
			},
			"secret_key": schema.StringAttribute{
				Description: "The secret key for S3-compatible API authentication. Only returned on create.",
				Computed:    true,
				Sensitive:   true,
			},
			"bucket_name": schema.StringAttribute{
				Description: "The bucket name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the object storage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the object storage was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the object storage was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *objectStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.SevallaClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.SevallaClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *objectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ObjectStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildCreateRequest(&plan)

	os, err := r.client.CreateObjectStorage(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Object Storage",
			"Could not create object storage, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenObjectStorage(ctx, os, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ObjectStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	os, err := r.client.GetObjectStorage(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Object Storage",
			"Could not read object storage ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenObjectStorage(ctx, os, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *objectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ObjectStorageResourceModel
	var state ObjectStorageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildUpdateRequest(&plan, &state)

	_, err := r.client.UpdateObjectStorage(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Object Storage",
			"Could not update object storage ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// PATCH response omits bucket_name, endpoint, access_key, secret_key.
	// Do a full GET to populate all fields.
	os, err := r.client.GetObjectStorage(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Object Storage After Update",
			"Could not read object storage ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenObjectStorage(ctx, os, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *objectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ObjectStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteObjectStorage(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Object Storage",
			"Could not delete object storage ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *objectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
