package application

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationResource{}
	_ resource.ResourceWithConfigure   = &applicationResource{}
	_ resource.ResourceWithImportState = &applicationResource{}
)

type applicationResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &applicationResource{}
}

func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the application (2-64 characters).",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(2, 64)},
			},
			"cluster_id": schema.StringAttribute{
				Description: "The cluster where the application is deployed. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Description: "The source type of the application. Valid values: `privateGit`, `publicGit`, `dockerImage`.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf("privateGit", "publicGit", "dockerImage")},
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the project to group the application under. Changing it moves the application to the new project; removing it from the configuration leaves the application in its current project.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_type": schema.StringAttribute{
				Description: "The Git provider type. Valid values: `github`, `bitbucket`, `gitlab`.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("github", "bitbucket", "gitlab")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_url": schema.StringAttribute{
				Description: "The repository URL for the application source.",
				Optional:    true,
			},
			"default_branch": schema.StringAttribute{
				Description: "The default branch to deploy from. Defaults to the repository's default branch.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"docker_image": schema.StringAttribute{
				Description: "The Docker image to deploy.",
				Optional:    true,
			},
			"docker_registry_credential_id": schema.StringAttribute{
				Description: "The ID of the Docker registry credential to use.",
				Optional:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Whether to automatically deploy on push.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"build_type": schema.StringAttribute{
				Description: "The build type. Valid values: `dockerfile`, `pack` (Cloud Native Buildpacks), `nixpacks`, `railpack`. Defaults to `dockerfile`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("dockerfile"),
				Validators:  []validator.String{stringvalidator.OneOf("dockerfile", "pack", "nixpacks", "railpack")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_path": schema.StringAttribute{
				Description: "The build path within the repository.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("."),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"build_cache_enabled": schema.BoolAttribute{
				Description: "Whether the build cache is enabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"git_lfs_enabled": schema.BoolAttribute{
				Description: "Whether Git LFS objects are fetched during the source checkout step. Enabled by default.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"hibernation_enabled": schema.BoolAttribute{
				Description: "Whether hibernation is enabled for the application.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"hibernate_after_seconds": schema.Int64Attribute{
				Description: "Number of seconds of inactivity before the application hibernates (1-604800).",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 604800)},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dockerfile_path": schema.StringAttribute{
				Description: "The path to the Dockerfile when `build_type` is `dockerfile`. Defaults to `Dockerfile`. Null for other build types.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					nullUnlessBuildType("dockerfile", "Dockerfile"),
				},
			},
			"docker_context": schema.StringAttribute{
				Description: "The Docker build context path when `build_type` is `dockerfile`. Defaults to `.`. Null for other build types.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					nullUnlessBuildType("dockerfile", "."),
				},
			},
			"pack_builder": schema.StringAttribute{
				Description: "The Cloud Native Buildpacks builder image when `build_type` is `pack`. Valid values: `heroku/builder:22`, `heroku/builder:24`.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("heroku/builder:22", "heroku/builder:24")},
				PlanModifiers: []planmodifier.String{
					nullUnlessBuildType("pack", ""),
				},
			},
			"nixpacks_version": schema.StringAttribute{
				Description: "The Nixpacks version to pin when `build_type` is `nixpacks`. Valid values: `1.29`, `1.39`, `1.41`. Defaults to the platform version.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("1.29", "1.39", "1.41")},
				PlanModifiers: []planmodifier.String{
					nullUnlessBuildType("nixpacks", ""),
				},
			},
			"allow_deploy_paths": schema.ListAttribute{
				Description: "Glob patterns of paths that trigger an automatic deployment when changed. Cannot be combined with a non-empty `ignore_deploy_paths`. Set to `[]` to clear.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_deploy_paths": schema.ListAttribute{
				Description: "Glob patterns of paths that never trigger an automatic deployment. Cannot be combined with a non-empty `allow_deploy_paths`. Set to `[]` to clear.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"buildpacks": schema.ListNestedAttribute{
				Description: "Custom buildpacks when `build_type` is `pack`. Null for other build types.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				PlanModifiers: []planmodifier.List{
					nullUnlessBuildType("pack", ""),
				},
				NestedObject: schema.NestedAttributeObject{
					// order and source are required, but Optional+Computed so that
					// buildpacks set outside Terraform do not cause a perpetual diff.
					Validators: []validator.Object{
						objectvalidator.AlsoRequires(path.MatchRelative().AtName("order"), path.MatchRelative().AtName("source")),
					},
					Attributes: map[string]schema.Attribute{
						"order": schema.Int64Attribute{
							Description: "The position of the buildpack in the build (0-20). Required.",
							Optional:    true,
							Computed:    true,
							Validators:  []validator.Int64{int64validator.Between(0, 20)},
						},
						"source": schema.StringAttribute{
							Description: "The buildpack ID or URL. Required.",
							Optional:    true,
							Computed:    true,
							Validators:  []validator.String{stringvalidator.LengthBetween(1, 500)},
						},
					},
				},
			},
			"cdn_enabled": schema.BoolAttribute{
				Description: "Whether the CDN is enabled for the application.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"edge_cache_enabled": schema.BoolAttribute{
				Description: "Whether edge caching is enabled for the application.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_for_checks": schema.BoolAttribute{
				Description:        "Has no effect.",
				DeprecationMessage: "The Sevalla API does not support wait_for_checks and the provider never sent it. Remove it from your configuration; it will be removed in a future major version.",
				Optional:           true,
				Computed:           true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			// Computed-only attributes
			"name": schema.StringAttribute{
				Description: "The system-generated name of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "The Kubernetes namespace of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The application type (app or previewApp).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the application.",
				Computed:    true,
			},
			"is_suspended": schema.BoolAttribute{
				Description: "Whether the application is currently suspended.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The user who created the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the application was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the application was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CreateApplication(ctx, buildCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Application",
			"Could not create application, unexpected error: "+err.Error(),
		)
		return
	}
	appID := app.ID
	// Save the ID first so a failure below leaves a tainted resource instead of an orphan.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), appID)...)

	// The create endpoint only accepts a limited set of fields. Apply any
	// additional plan values (e.g. auto_deploy, build_type) via an update.
	if updateReq := buildPostCreateUpdateRequest(ctx, &plan); updateReq != nil {
		if _, err := r.client.UpdateApplication(ctx, appID, updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Application After Create",
				"Application was created but failed to apply additional settings: "+err.Error(),
			)
			return
		}
	}

	if err := r.applyToggles(ctx, appID, &plan, nil); err != nil {
		resp.Diagnostics.AddError("Error Updating Application After Create", err.Error())
		return
	}

	r.readInto(ctx, appID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read application ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenApplication(ctx, app, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	var state ApplicationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	appID := state.ID.ValueString()

	if updateReq := buildUpdateRequest(ctx, &plan, &state); !reflect.ValueOf(*updateReq).IsZero() {
		if _, err := r.client.UpdateApplication(ctx, appID, updateReq); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Application",
				"Could not update application ID "+appID+": "+err.Error(),
			)
			return
		}
	}

	if !plan.ProjectID.IsUnknown() && !plan.ProjectID.IsNull() && !plan.ProjectID.Equal(state.ProjectID) {
		err := r.client.AddProjectService(ctx, plan.ProjectID.ValueString(), &client.AddProjectServiceRequest{
			ServiceID:   appID,
			ServiceType: "app",
		})
		if err != nil {
			resp.Diagnostics.AddError("Error Updating Application", "Could not move application ID "+appID+" to project: "+err.Error())
			return
		}
	}

	if err := r.applyToggles(ctx, appID, &plan, &state); err != nil {
		resp.Diagnostics.AddError("Error Updating Application", err.Error())
		return
	}

	// The PATCH response omits fields such as buildpacks, so read the full application.
	r.readInto(ctx, appID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// applyToggles flips the CDN and edge cache settings when the plan changes
// them (state is nil on create) and the application's current value differs.
func (r *applicationResource) applyToggles(ctx context.Context, appID string, plan, state *ApplicationResourceModel) error {
	type setting struct {
		name   string
		want   types.Bool
		prior  types.Bool
		get    func(*client.Application) *bool
		toggle func(context.Context, string) (bool, error)
	}
	settings := []setting{
		{"the CDN", plan.CDNEnabled, types.BoolNull(), func(a *client.Application) *bool { return a.IsCDNEnabled }, r.client.ToggleApplicationCDN},
		{"edge caching", plan.EdgeCacheEnabled, types.BoolNull(), func(a *client.Application) *bool { return a.IsEdgeCacheEnabled }, r.client.ToggleApplicationEdgeCache},
	}
	if state != nil {
		settings[0].prior = state.CDNEnabled
		settings[1].prior = state.EdgeCacheEnabled
	}

	var app *client.Application
	for _, s := range settings {
		if s.want.IsUnknown() || s.want.IsNull() || s.want.Equal(s.prior) {
			continue
		}
		if app == nil {
			var err error
			if app, err = r.client.GetApplication(ctx, appID); err != nil {
				return fmt.Errorf("could not read application ID %s: %w", appID, err)
			}
		}
		current := s.get(app)
		if current == nil {
			return fmt.Errorf("the status of %s of application ID %s is temporarily unavailable; try again later", s.name, appID)
		}
		if *current == s.want.ValueBool() {
			continue
		}
		on, err := s.toggle(ctx, appID)
		if err != nil {
			return fmt.Errorf("could not toggle %s of application ID %s: %w", s.name, appID, err)
		}
		if on != s.want.ValueBool() {
			return fmt.Errorf("toggling %s of application ID %s did not change it", s.name, appID)
		}
	}
	return nil
}

func (r *applicationResource) readInto(ctx context.Context, appID string, model *ApplicationResourceModel, diags *diag.Diagnostics) {
	app, err := r.client.GetApplication(ctx, appID)
	if err != nil {
		diags.AddError("Error Reading Application", "Could not read application ID "+appID+": "+err.Error())
		return
	}
	diags.Append(flattenApplication(ctx, app, model)...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting Application",
			"Could not delete application ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// nullUnlessBuildTypeModifier is a plan modifier that sets a string attribute
// to null when build_type does not match the required type, because
// build-type-specific fields (e.g. nixpacks_version, dockerfile_path) have no
// effect for other build types and are stored as null in state. It also
// applies defaultValue when the build type is first set, which a schema
// Default cannot do without planning an update for every other build type.
type nullUnlessBuildTypeModifier struct {
	requiredBuildType string
	defaultValue      string
}

func nullUnlessBuildType(bt, defaultValue string) nullUnlessBuildTypeModifier {
	return nullUnlessBuildTypeModifier{requiredBuildType: bt, defaultValue: defaultValue}
}

func (m nullUnlessBuildTypeModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Set to null when build_type is not %q", m.requiredBuildType)
}

func (m nullUnlessBuildTypeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// planAction decides what an unconfigured or irrelevant value is planned as.
type planAction int

const (
	planKeep    planAction = iota // keep the planned value
	planNull                      // the build type does not use the attribute
	planUnknown                   // the API decides the value
	planState                     // carry the value forward from state
	planDefault                   // use defaultValue
)

func (m nullUnlessBuildTypeModifier) action(ctx context.Context, plan tfsdk.Plan, state tfsdk.State, configNull bool, diags *diag.Diagnostics) planAction {
	var buildType types.String
	diags.Append(plan.GetAttribute(ctx, path.Root("build_type"), &buildType)...)
	switch {
	case diags.HasError():
		return planKeep
	case buildType.IsUnknown():
		return planUnknown
	case buildType.ValueString() != m.requiredBuildType:
		return planNull
	case !configNull:
		return planKeep
	}

	// The build type matches but the value is not configured. Keep it if the
	// build type is unchanged; otherwise use the default, or let the API decide.
	var stateBuildType types.String
	if !state.Raw.IsNull() {
		diags.Append(state.GetAttribute(ctx, path.Root("build_type"), &stateBuildType)...)
	}
	switch {
	case stateBuildType.ValueString() == m.requiredBuildType:
		return planState
	case m.defaultValue != "":
		return planDefault
	default:
		return planUnknown
	}
}

func (m nullUnlessBuildTypeModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	switch m.action(ctx, req.Plan, req.State, req.ConfigValue.IsNull(), &resp.Diagnostics) {
	case planNull:
		resp.PlanValue = types.StringNull()
	case planUnknown:
		resp.PlanValue = types.StringUnknown()
	case planState:
		resp.PlanValue = req.StateValue
	case planDefault:
		resp.PlanValue = types.StringValue(m.defaultValue)
	}
}

func (m nullUnlessBuildTypeModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	switch m.action(ctx, req.Plan, req.State, req.ConfigValue.IsNull(), &resp.Diagnostics) {
	case planNull:
		resp.PlanValue = types.ListNull(req.PlanValue.ElementType(ctx))
	case planUnknown:
		resp.PlanValue = types.ListUnknown(req.PlanValue.ElementType(ctx))
	case planState:
		resp.PlanValue = req.StateValue
	}
}
