package staticsite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &staticSiteResource{}
	_ resource.ResourceWithConfigure   = &staticSiteResource{}
	_ resource.ResourceWithImportState = &staticSiteResource{}
)

type staticSiteResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &staticSiteResource{}
}

func (r *staticSiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site"
}

func (r *staticSiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla static site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the static site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the static site.",
				Required:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source type of the static site. Valid values: privateGit, publicGit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("publicGit"),
			},
			"git_type": schema.StringAttribute{
				Description: "The Git provider type. Valid values: github, bitbucket, gitlab.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_url": schema.StringAttribute{
				Description: "The repository URL for the static site source.",
				Required:    true,
			},
			"default_branch": schema.StringAttribute{
				Description: "The default branch to deploy from.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID to associate with this static site.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Whether to automatically deploy on push.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_preview_enabled": schema.BoolAttribute{
				Description: "Whether preview deployments are enabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"install_command": schema.StringAttribute{
				Description: "The install command to run.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_command": schema.StringAttribute{
				Description: "The build command to run.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_directory": schema.StringAttribute{
				Description: "The directory where the built static files are published.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"root_directory": schema.StringAttribute{
				Description: "The root directory within the repository.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"node_version": schema.StringAttribute{
				Description: "The Node.js version to use for building.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index_file": schema.StringAttribute{
				Description: "The index file for the static site.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"error_file": schema.StringAttribute{
				Description: "The error file for the static site.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// Computed-only attributes
			"name": schema.StringAttribute{
				Description: "The system-generated name of the static site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the static site.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the static site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the static site.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the static site was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the static site was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *staticSiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *staticSiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticSiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildCreateRequest(&plan)

	createResult, err := r.client.CreateStaticSite(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Static Site",
			"Could not create static site, unexpected error: "+err.Error(),
		)
		return
	}

	// Create response may be incomplete — do a GET to retrieve the full object.
	ss, err := r.client.GetStaticSite(ctx, createResult.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Static Site After Create",
			"Could not read static site ID "+createResult.ID+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenStaticSite(ctx, ss, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticSiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticSiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ss, err := r.client.GetStaticSite(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Static Site",
			"Could not read static site ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenStaticSite(ctx, ss, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *staticSiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StaticSiteResourceModel
	var state StaticSiteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ssID := state.ID.ValueString()
	updateReq := buildUpdateRequest(ctx, &plan, &state)

	if _, err := r.client.UpdateStaticSite(ctx, ssID, updateReq); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Static Site",
			"Could not update static site ID "+ssID+": "+err.Error(),
		)
		return
	}

	// Update response may be incomplete — do a GET to retrieve the full object.
	ss, err := r.client.GetStaticSite(ctx, ssID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Static Site After Update",
			"Could not read static site ID "+ssID+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenStaticSite(ctx, ss, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticSiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticSiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStaticSite(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Static Site",
			"Could not delete static site ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *staticSiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
