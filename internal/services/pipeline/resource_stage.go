package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &pipelineStageResource{}
	_ resource.ResourceWithConfigure   = &pipelineStageResource{}
	_ resource.ResourceWithImportState = &pipelineStageResource{}
)

type pipelineStageResource struct {
	client *client.SevallaClient
}

type PipelineStageResourceModel struct {
	ID              types.String `tfsdk:"id"`
	PipelineID      types.String `tfsdk:"pipeline_id"`
	Name            types.String `tfsdk:"name"`
	InsertBefore    types.Int64  `tfsdk:"insert_before"`
	Type            types.String `tfsdk:"type"`
	Order           types.Int64  `tfsdk:"order"`
	Branch          types.String `tfsdk:"branch"`
	AutoCreateApp   types.Bool   `tfsdk:"auto_create_app"`
	DeleteStaleApps types.Bool   `tfsdk:"delete_stale_apps"`
	StaleAppDays    types.Int64  `tfsdk:"stale_app_days"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func NewStageResource() resource.Resource {
	return &pipelineStageResource{}
}

func (r *pipelineStageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_stage"
}

func (r *pipelineStageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a stage within a Sevalla pipeline.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the pipeline stage.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pipeline_id": schema.StringAttribute{
				Description: "The ID of the pipeline this stage belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the pipeline stage (1-255 characters).",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"insert_before": schema.Int64Attribute{
				Description: "Position to insert the stage at (0-50). Existing stages at this position and above are shifted up. Changing it creates a new stage, except after an import, when it is only recorded.",
				Required:    true,
				Validators:  []validator.Int64{int64validator.Between(0, 50)},
				PlanModifiers: []planmodifier.Int64{
					// The API does not return insert_before, so an imported stage has none.
					int64planmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.StateValue.IsNull()
					}, "Changing insert_before creates a new stage.", "Changing `insert_before` creates a new stage."),
				},
			},
			"type": schema.StringAttribute{
				Description: "The stage type.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"order": schema.Int64Attribute{
				Description: "The current position of the stage, starting from 0. It changes when other stages are added or removed.",
				Computed:    true,
			},
			"branch": schema.StringAttribute{
				Description: "The Git branch of the stage, used in branch-type pipelines.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_create_app": schema.BoolAttribute{
				Description: "Whether to automatically create applications when new branches are detected.",
				Computed:    true,
			},
			"delete_stale_apps": schema.BoolAttribute{
				Description: "Whether to automatically delete applications when their branches are removed.",
				Computed:    true,
			},
			"stale_app_days": schema.Int64Attribute{
				Description: "Number of days after which an application is considered stale and eligible for deletion.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the stage was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the stage was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *pipelineStageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pipelineStageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PipelineStageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipelineID := plan.PipelineID.ValueString()
	createReq := &client.CreatePipelineStageRequest{
		DisplayName:  plan.Name.ValueString(),
		InsertBefore: int(plan.InsertBefore.ValueInt64()),
	}
	if !plan.Branch.IsNull() && !plan.Branch.IsUnknown() {
		v := plan.Branch.ValueString()
		createReq.Branch = &v
	}

	stage, err := r.client.CreatePipelineStage(ctx, pipelineID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Pipeline Stage", err.Error())
		return
	}

	flattenPipelineStage(stage, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineStageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PipelineStageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipeline, err := r.client.GetPipeline(ctx, state.PipelineID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Pipeline Stage", err.Error())
		return
	}

	var found *client.PipelineStage
	for i := range pipeline.Stages {
		if pipeline.Stages[i].ID == state.ID.ValueString() {
			found = &pipeline.Stages[i]
			break
		}
	}

	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenPipelineStage(found, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is only reached when insert_before is set on an imported stage; there is nothing to send.
func (r *pipelineStageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PipelineStageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.InsertBefore = plan.InsertBefore
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pipelineStageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PipelineStageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePipelineStage(ctx, state.PipelineID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Pipeline Stage", err.Error())
		return
	}
}

func (r *pipelineStageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'pipeline_id/stage_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pipeline_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenPipelineStage(s *client.PipelineStage, model *PipelineStageResourceModel) {
	model.ID = types.StringValue(s.ID)
	model.Name = types.StringValue(s.DisplayName)
	model.Type = types.StringValue(s.Type)
	model.Order = types.Int64Value(int64(s.Order))
	if s.Branch != nil {
		model.Branch = types.StringValue(*s.Branch)
	} else {
		model.Branch = types.StringNull()
	}
	if s.AutoCreateApp != nil {
		model.AutoCreateApp = types.BoolValue(*s.AutoCreateApp)
	} else {
		model.AutoCreateApp = types.BoolNull()
	}
	if s.DeleteStaleApps != nil {
		model.DeleteStaleApps = types.BoolValue(*s.DeleteStaleApps)
	} else {
		model.DeleteStaleApps = types.BoolNull()
	}
	if s.StaleAppDays != nil {
		model.StaleAppDays = types.Int64Value(int64(*s.StaleAppDays))
	} else {
		model.StaleAppDays = types.Int64Null()
	}
	if s.CreatedAt != nil {
		model.CreatedAt = types.StringValue(*s.CreatedAt)
	}
	if s.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(*s.UpdatedAt)
	}
}
