package pipeline

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
	_ resource.Resource                = &pipelineStageApplicationResource{}
	_ resource.ResourceWithConfigure   = &pipelineStageApplicationResource{}
	_ resource.ResourceWithImportState = &pipelineStageApplicationResource{}
)

type pipelineStageApplicationResource struct {
	client *client.SevallaClient
}

type PipelineStageApplicationResourceModel struct {
	PipelineID    types.String `tfsdk:"pipeline_id"`
	StageID       types.String `tfsdk:"stage_id"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func NewStageApplicationResource() resource.Resource {
	return &pipelineStageApplicationResource{}
}

func (r *pipelineStageApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_stage_application"
}

func (r *pipelineStageApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches an application to a pipeline stage in Sevalla.",
		Attributes: map[string]schema.Attribute{
			"pipeline_id": schema.StringAttribute{
				Description: "The ID of the pipeline.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"stage_id": schema.StringAttribute{
				Description: "The ID of the pipeline stage.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application to attach to the stage.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *pipelineStageApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pipelineStageApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PipelineStageApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddPipelineStageApp(ctx, plan.PipelineID.ValueString(), plan.StageID.ValueString(), &client.AddPipelineStageAppRequest{
		ApplicationID: plan.ApplicationID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Pipeline Stage Application", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pipelineStageApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PipelineStageApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the pipeline still exists; if not, remove this resource from state.
	pipeline, err := r.client.GetPipeline(ctx, state.PipelineID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Pipeline Stage Application", err.Error())
		return
	}

	// Verify the stage still exists within the pipeline.
	stageFound := false
	for _, stage := range pipeline.Stages {
		if stage.ID == state.StageID.ValueString() {
			stageFound = true
			break
		}
	}

	if !stageFound {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pipelineStageApplicationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Pipeline stage application attachments cannot be updated. Delete and recreate the resource instead.",
	)
}

func (r *pipelineStageApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PipelineStageApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemovePipelineStageApp(ctx, state.PipelineID.ValueString(), state.StageID.ValueString(), state.ApplicationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Pipeline Stage Application", err.Error())
		return
	}
}

func (r *pipelineStageApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'pipeline_id/stage_id/application_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pipeline_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stage_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[2])...)
}
