package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

const databaseReadyTimeout = 5 * time.Minute

var (
	_ resource.Resource                = &databaseResource{}
	_ resource.ResourceWithConfigure   = &databaseResource{}
	_ resource.ResourceWithImportState = &databaseResource{}
)

type databaseResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &databaseResource{}
}

func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *databaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the database.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The database engine type. Valid values: postgresql, mysql, redis, mariadb, valkey. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The database engine version. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The cluster where the database is deployed. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The resource type (size) of the database. Can be changed to resize.",
				Required:    true,
			},
			"db_name": schema.StringAttribute{
				Description: "The database name. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_password": schema.StringAttribute{
				Description: "The database password. Cannot be changed after creation.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_user": schema.StringAttribute{
				Description: "The database user. Cannot be \"root\" or \"postgres\". If omitted, the database engine default user is used. Cannot be changed after creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID to associate with this database.",
				Optional:    true,
			},
			"extensions": schema.SingleNestedAttribute{
				Description: "PostgreSQL extensions to enable. Only applicable when type is postgresql. Cannot be changed after creation.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"enable_vector": schema.BoolAttribute{
						Description: "Enable the pgvector extension for vector similarity search.",
						Optional:    true,
					},
					"enable_postgis": schema.BoolAttribute{
						Description: "Enable the PostGIS extension for geospatial data.",
						Optional:    true,
					},
					"enable_cron": schema.BoolAttribute{
						Description: "Enable the pg_cron extension for scheduled jobs.",
						Optional:    true,
					},
				},
			},
			"name": schema.StringAttribute{
				Description: "The system-generated name of the database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the database.",
				Computed:    true,
			},
			"is_suspended": schema.BoolAttribute{
				Description: "Whether the database is currently suspended.",
				Computed:    true,
			},
			"cluster_display_name": schema.StringAttribute{
				Description: "The display name of the cluster.",
				Computed:    true,
			},
			"cluster_location": schema.StringAttribute{
				Description: "The location of the cluster.",
				Computed:    true,
			},
			"resource_type_name": schema.StringAttribute{
				Description: "The name of the resource type.",
				Computed:    true,
			},
			"cpu_limit": schema.Int64Attribute{
				Description: "The CPU limit for the database.",
				Computed:    true,
			},
			"memory_limit": schema.Int64Attribute{
				Description: "The memory limit for the database in bytes.",
				Computed:    true,
			},
			"storage_size": schema.Int64Attribute{
				Description: "The storage size for the database in bytes.",
				Computed:    true,
			},
			"internal_hostname": schema.StringAttribute{
				Description: "The internal hostname for connecting to the database.",
				Computed:    true,
			},
			"internal_port": schema.StringAttribute{
				Description: "The internal port for connecting to the database.",
				Computed:    true,
			},
			"external_hostname": schema.StringAttribute{
				Description: "The external hostname for connecting to the database.",
				Computed:    true,
				Sensitive:   true,
			},
			"external_port": schema.StringAttribute{
				Description: "The external port for connecting to the database.",
				Computed:    true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the database was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the database was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *databaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildCreateRequest(&plan)

	createResult, err := r.client.CreateDatabase(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Database",
			"Could not create database, unexpected error: "+err.Error(),
		)
		return
	}

	// Wait for the database to reach ready status before reading.
	db, err := r.client.WaitForDatabaseStatus(ctx, createResult.ID, []string{"ready"}, databaseReadyTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Database Ready",
			"Database "+createResult.ID+" did not become ready: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDatabase(ctx, db, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.GetDatabase(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Database",
			"Could not read database ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDatabase(ctx, db, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DatabaseResourceModel
	var state DatabaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := state.ID.ValueString()

	// Database must be in a ready state before it can be updated.
	_, err := r.client.WaitForDatabaseStatus(ctx, dbID, []string{"ready"}, databaseReadyTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Waiting for Database Ready Before Update",
			"Database "+dbID+" did not become ready: "+err.Error(),
		)
		return
	}

	updateReq := buildUpdateRequest(&plan, &state)

	_, err = r.client.UpdateDatabase(ctx, dbID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Database",
			"Could not update database ID "+dbID+": "+err.Error(),
		)
		return
	}

	// Update response may be incomplete — do a GET to retrieve the full object.
	db, err := r.client.GetDatabase(ctx, dbID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Database After Update",
			"Could not read database ID "+dbID+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenDatabase(ctx, db, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := state.ID.ValueString()

	// Database must be in a ready state before it can be deleted.
	_, err := r.client.WaitForDatabaseStatus(ctx, dbID, []string{"ready", "error", "suspended"}, databaseReadyTimeout)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Waiting for Database Ready Before Delete",
			"Database "+dbID+" did not reach a deletable state: "+err.Error(),
		)
		return
	}

	err = r.client.DeleteDatabase(ctx, dbID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Database",
			"Could not delete database ID "+dbID+": "+err.Error(),
		)
		return
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
