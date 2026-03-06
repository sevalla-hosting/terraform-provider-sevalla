package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &databaseInternalConnectionResource{}
	_ resource.ResourceWithConfigure   = &databaseInternalConnectionResource{}
	_ resource.ResourceWithImportState = &databaseInternalConnectionResource{}
)

type databaseInternalConnectionResource struct {
	client *client.SevallaClient
}

type DatabaseInternalConnectionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	DatabaseID        types.String `tfsdk:"database_id"`
	TargetID          types.String `tfsdk:"target_id"`
	TargetType        types.String `tfsdk:"target_type"`
	SourceType        types.String `tfsdk:"source_type"`
	SourceDisplayName types.String `tfsdk:"source_display_name"`
	TargetDisplayName types.String `tfsdk:"target_display_name"`
}

func NewInternalConnectionResource() resource.Resource {
	return &databaseInternalConnectionResource{}
}

func (r *databaseInternalConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_internal_connection"
}

func (r *databaseInternalConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an internal connection between a Sevalla database and another resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the internal connection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"database_id": schema.StringAttribute{
				Description: "The ID of the source database.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Description: "The ID of the target resource to connect.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "The type of the target resource. Valid values: app, database.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_type": schema.StringAttribute{
				Description: "The type of the source resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_display_name": schema.StringAttribute{
				Description: "The display name of the source resource.",
				Computed:    true,
			},
			"target_display_name": schema.StringAttribute{
				Description: "The display name of the target resource.",
				Computed:    true,
			},
		},
	}
}

func (r *databaseInternalConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseInternalConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseInternalConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := plan.DatabaseID.ValueString()
	createReq := &client.CreateInternalConnectionRequest{
		TargetID:   plan.TargetID.ValueString(),
		TargetType: plan.TargetType.ValueString(),
	}

	conn, err := r.client.CreateInternalConnection(ctx, dbID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Database Internal Connection", err.Error())
		return
	}

	plan.ID = types.StringValue(conn.ID)

	// Create response only returns ID, so read to get all fields.
	connections, err := r.client.ListInternalConnections(ctx, dbID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Database Internal Connection After Create", err.Error())
		return
	}

	var found *client.InternalConnection
	for i := range connections {
		if connections[i].ID == conn.ID {
			found = &connections[i]
			break
		}
	}

	if found != nil {
		flattenInternalConnection(found, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseInternalConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabaseInternalConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbID := state.DatabaseID.ValueString()
	connections, err := r.client.ListInternalConnections(ctx, dbID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Database Internal Connection", err.Error())
		return
	}

	var found *client.InternalConnection
	for i := range connections {
		if connections[i].ID == state.ID.ValueString() {
			found = &connections[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenInternalConnection(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseInternalConnectionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Database internal connections cannot be updated. Delete and recreate the resource instead.",
	)
}

func (r *databaseInternalConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabaseInternalConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteInternalConnection(ctx, state.DatabaseID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Database Internal Connection", err.Error())
		return
	}
}

func (r *databaseInternalConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'database_id/connection_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("database_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// normalizeTargetType maps the read-side enum values (appResource, dbResource)
// back to the create-side values (app, database) to avoid perpetual diffs.
func normalizeTargetType(apiValue string) string {
	switch apiValue {
	case "appResource":
		return "app"
	case "dbResource":
		return "database"
	default:
		return apiValue
	}
}

func flattenInternalConnection(conn *client.InternalConnection, model *DatabaseInternalConnectionResourceModel) {
	model.ID = types.StringValue(conn.ID)
	model.TargetID = types.StringValue(conn.TargetID)

	if conn.SourceType != nil {
		model.SourceType = types.StringValue(*conn.SourceType)
	} else {
		model.SourceType = types.StringNull()
	}
	if conn.TargetType != nil {
		model.TargetType = types.StringValue(normalizeTargetType(*conn.TargetType))
	} else {
		model.TargetType = types.StringNull()
	}
	if conn.SourceDisplayName != nil {
		model.SourceDisplayName = types.StringValue(*conn.SourceDisplayName)
	} else {
		model.SourceDisplayName = types.StringNull()
	}
	if conn.TargetDisplayName != nil {
		model.TargetDisplayName = types.StringValue(*conn.TargetDisplayName)
	} else {
		model.TargetDisplayName = types.StringNull()
	}
}
