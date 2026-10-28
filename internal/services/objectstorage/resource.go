package objectstorage

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sevalla/terraform-provider-sevalla/internal/client"
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
				Description: "The display name of the object storage (2-64 characters).",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(2, 64)},
			},
			"location": schema.StringAttribute{
				Description: "Geographic hint for where most data access occurs. Valid values: `apac`, `eeur`, `enam`, `oc`, `weur`, `wnam`. Defaults to `enam`. Changing this creates a new bucket.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("apac", "eeur", "enam", "oc", "weur", "wnam")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"jurisdiction": schema.StringAttribute{
				Description: "Data residency jurisdiction. Valid values: `default`, `eu`, `fedramp`. Defaults to `default`. Changing this creates a new bucket.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("default", "eu", "fedramp")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_access": schema.BoolAttribute{
				Description: "Whether the bucket has a public CDN domain. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to group the bucket under. Removing it removes the bucket from the project.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "The lifecycle status of the bucket: `active`, `deleting` or `deletionFailed`.",
				Computed:    true,
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_key": schema.StringAttribute{
				Description: "The access key ID for the S3-compatible API. The API only returns it when the bucket is created, so it is null after an import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_key": schema.StringAttribute{
				Description: "The secret access key for the S3-compatible API. The API only returns it when the bucket is created, so it is null after an import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	id := state.ID.ValueString()
	if updateReq := buildUpdateRequest(&plan, &state); !reflect.ValueOf(*updateReq).IsZero() {
		if _, err := r.client.UpdateObjectStorage(ctx, id, updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Object Storage",
				"Could not update object storage ID "+id+": "+err.Error(),
			)
			return
		}
	}

	if !plan.PublicAccess.IsUnknown() && !plan.PublicAccess.Equal(state.PublicAccess) {
		var err error
		if plan.PublicAccess.ValueBool() {
			err = r.client.EnableObjectStorageCDN(ctx, id)
		} else {
			err = r.client.DisableObjectStorageCDN(ctx, id)
		}
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Object Storage", "Could not change public access of object storage ID "+id+": "+err.Error())
			return
		}
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

	id := state.ID.ValueString()
	err := r.client.DeleteObjectStorage(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Object Storage",
			"Could not delete object storage ID "+id+": "+err.Error(),
		)
		return
	}

	// Deletion empties and removes the bucket in the background.
	err = client.Poll(ctx, deleteTimeout, func() (bool, error) {
		os, err := r.client.GetObjectStorage(ctx, id)
		if client.IsNotFound(err) {
			return true, nil
		}
		if err == nil && os.Status == "deletionFailed" {
			err = fmt.Errorf("deletion of object storage ID %s failed", id)
		}
		return false, err
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for Object Storage Deletion", err.Error())
	}
}

// deleteTimeout bounds how long Delete waits for the bucket to be removed.
const deleteTimeout = 30 * time.Minute

func (r *objectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
