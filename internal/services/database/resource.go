package database

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

const databaseReadyTimeout = 15 * time.Minute

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
				Description: "The display name of the database (2-64 characters).",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(2, 64)},
			},
			"type": schema.StringAttribute{
				Description: "The database engine type. Valid values: `postgresql`, `mysql`, `mariadb`, `redis`, `valkey`. Changing this creates a new database.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("postgresql", "mysql", "mariadb", "redis", "valkey")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The database engine version. The available versions depend on the type and cluster. Changing this creates a new database.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The cluster where the database is deployed. Changing this creates a new database.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The resource type (size) of the database. Changing it resizes the database in place; on OCI clusters, resizing a DB1 database deletes its backups.",
				Required:    true,
			},
			"db_name": schema.StringAttribute{
				Description: "The name of the database to create (2-100 characters). Redis and Valkey ignore it. Changing this creates a new database, except for Redis and Valkey.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(2, 100)},
				PlanModifiers: []planmodifier.String{
					requiresReplaceUnlessKeyValue(),
				},
			},
			"db_password": schema.StringAttribute{
				Description: "The password of the database user (4-100 characters). Changing it resets the password in place. The API never returns it, so after an import the configured value is recorded without being applied.",
				Required:    true,
				Sensitive:   true,
				Validators:  []validator.String{stringvalidator.LengthBetween(4, 100)},
			},
			"db_user": schema.StringAttribute{
				Description: "The database user (2-32 characters, not `root` or `postgres`). Defaults to the engine's default user. Redis and Valkey ignore it. Changing this creates a new database, except for Redis and Valkey.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(2, 32)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					requiresReplaceUnlessKeyValue(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to group the database under. Removing it removes the database from the project.",
				Optional:    true,
			},
			"extensions": schema.SingleNestedAttribute{
				Description: "PostgreSQL extensions to enable (PostgreSQL only). They can only be set when the database is created and the API does not return them, so they are not checked for drift. Changing or removing them creates a new database; setting them where none were recorded (e.g. after an import) only records them.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
						// The API does not return extensions, so a null state (e.g. after an
						// import) is not a known value to compare against.
						resp.RequiresReplace = !req.StateValue.IsNull()
					}, "Changing extensions creates a new database.", "Changing `extensions` creates a new database."),
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
				Description: "The CPU limit for the database in millicores.",
				Computed:    true,
			},
			"memory_limit": schema.Int64Attribute{
				Description: "The memory limit for the database in megabytes.",
				Computed:    true,
			},
			"storage_size": schema.Int64Attribute{
				Description: "The storage size for the database in megabytes.",
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

	createResult, err := r.client.CreateDatabase(ctx, buildCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Database",
			"Could not create database, unexpected error: "+err.Error(),
		)
		return
	}
	// Save the ID first so a failure below leaves a tainted resource instead of an orphan.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), createResult.ID)...)

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
	updateReq := buildUpdateRequest(&plan, &state)
	// The API never returns the password, so a null state value (e.g. after an import) is only recorded.
	resetPassword := !state.DbPassword.IsNull() && !plan.DbPassword.Equal(state.DbPassword)

	if resetPassword || !reflect.ValueOf(*updateReq).IsZero() {
		// The database must be ready before it can be changed.
		if _, err := r.client.WaitForDatabaseStatus(ctx, dbID, []string{"ready"}, databaseReadyTimeout); err != nil {
			resp.Diagnostics.AddError(
				"Error Waiting for Database Ready Before Update",
				"Database "+dbID+" did not become ready: "+err.Error(),
			)
			return
		}
	}

	if !reflect.ValueOf(*updateReq).IsZero() {
		if _, err := r.client.UpdateDatabase(ctx, dbID, updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Database",
				"Could not update database ID "+dbID+": "+err.Error(),
			)
			return
		}
		// Resizing runs asynchronously.
		if _, err := r.client.WaitForDatabaseStatus(ctx, dbID, []string{"ready"}, databaseReadyTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Database Update", "Database "+dbID+" did not become ready: "+err.Error())
			return
		}
	}

	if resetPassword {
		if err := r.client.ResetDatabasePassword(ctx, dbID, plan.DbPassword.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error Updating Database", "Could not reset the password of database ID "+dbID+": "+err.Error())
			return
		}
		if _, err := r.client.WaitForDatabaseStatus(ctx, dbID, []string{"ready"}, databaseReadyTimeout); err != nil {
			resp.Diagnostics.AddError("Error Waiting for Database Password Reset", "Database "+dbID+" did not become ready: "+err.Error())
			return
		}
	}

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

	// Wait for any operation in progress to finish before deleting.
	_, err := r.client.WaitForDatabaseStatus(ctx, dbID, deletableStatuses, databaseReadyTimeout)
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
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Database",
			"Could not delete database ID "+dbID+": "+err.Error(),
		)
		return
	}

	err = client.Poll(ctx, databaseReadyTimeout, func() (bool, error) {
		db, err := r.client.GetDatabase(ctx, dbID)
		if client.IsNotFound(err) {
			return true, nil
		}
		if err == nil && db.Status != nil && *db.Status == "dbDeletionFailed" {
			err = fmt.Errorf("deletion of database %s failed", dbID)
		}
		return false, err
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Waiting for Database Deletion", err.Error())
	}
}

// requiresReplaceUnlessKeyValue forces replacement when the value changes,
// except for Redis and Valkey, which ignore db_name and db_user.
func requiresReplaceUnlessKeyValue() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		var dbType types.String
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("type"), &dbType)...)
		resp.RequiresReplace = !isKeyValueStore(dbType.ValueString())
	}, "Changing this creates a new database, except for Redis and Valkey.", "Changing this creates a new database, except for Redis and Valkey.")
}

func isKeyValueStore(dbType string) bool {
	return dbType == "redis" || dbType == "valkey"
}

// deletableStatuses are the statuses in which no operation is in progress.
var deletableStatuses = []string{"ready", "error", "passwordChangeFailed", "dbDeletionFailed", "restoringFailed"}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
