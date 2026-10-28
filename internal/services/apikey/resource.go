package apikey

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                     = &apiKeyResource{}
	_ resource.ResourceWithConfigure        = &apiKeyResource{}
	_ resource.ResourceWithImportState      = &apiKeyResource{}
	_ resource.ResourceWithConfigValidators = &apiKeyResource{}
)

type apiKeyResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &apiKeyResource{}
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the API key (1-255 characters).",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"expires_at": schema.StringAttribute{
				Description: "When the API key expires, as an RFC 3339 timestamp (e.g. `2027-01-01T00:00:00Z`). Omit for a key that never expires. Changing this creates a new API key.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					// The API returns the timestamp with milliseconds; the same instant is no change.
					stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !sameInstant(req.PlanValue.ValueString(), req.StateValue.ValueString())
					}, "Changing expires_at creates a new API key.", "Changing `expires_at` creates a new API key."),
				},
			},
			"token": schema.StringAttribute{
				Description: "The API key token. Only returned on create.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"masked_token": schema.StringAttribute{
				Description: "The token with all but its first 6 and last 4 characters masked.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the API key can be used to authenticate. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"capabilities": schema.ListNestedAttribute{
				Description: "Capabilities (permissions) to grant to the API key. At least one of `capabilities` or `role_ids` is required.",
				Optional:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							Description: "The permission in `RESOURCE:ACTION` format (e.g. `APP:READ`, `DATABASE:CREATE`). Use the `sevalla_api_key_permissions` data source to list them.",
							Required:    true,
							Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						},
						"id_resource": schema.StringAttribute{
							Description: "Optional resource ID to scope this permission to a specific resource.",
							Optional:    true,
						},
					},
				},
			},
			"role_ids": schema.ListAttribute{
				Description: "IDs of the roles to assign to the API key. Use the `sevalla_api_key_roles` data source to discover available roles. At least one of `capabilities` or `role_ids` is required.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
			},
			"roles": schema.ListNestedAttribute{
				Description: "The roles assigned to this API key.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the role.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the role.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the role.",
							Computed:    true,
						},
					},
				},
			},
			"source": schema.StringAttribute{
				Description: "How the API key was created: `DASHBOARD`, `CLI`, `EXTERNAL`, `MCP` or `SUPPORT_MCP`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Description: "The timestamp when the API key was last used for authentication.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the API key was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the API key was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apiKeyResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(path.MatchRoot("capabilities"), path.MatchRoot("role_ids")),
	}
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq, diags := buildCreateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAPIKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating API Key",
			"Could not create API key, unexpected error: "+err.Error(),
		)
		return
	}
	// The token is only returned here, so store it before anything else can fail.
	plan.ID = types.StringValue(created.ID)
	plan.Token = types.StringValue(created.Token)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), plan.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("token"), plan.Token)...)

	apiKey, err := r.client.GetAPIKey(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading API Key After Create",
			"API key was created but could not read it: "+err.Error(),
		)
		return
	}

	apiKey, err = r.setEnabled(ctx, apiKey, plan.Enabled)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating API Key", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenAPIKey(ctx, apiKey, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading API Key",
			"Could not read API key ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flattenAPIKey(ctx, apiKey, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan APIKeyResourceModel
	var state APIKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq, diags := buildUpdateRequest(ctx, &plan, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiKey *client.APIKey
	var err error
	if reflect.ValueOf(*updateReq).IsZero() {
		apiKey, err = r.client.GetAPIKey(ctx, state.ID.ValueString())
	} else {
		apiKey, err = r.client.UpdateAPIKey(ctx, state.ID.ValueString(), updateReq)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating API Key",
			"Could not update API key ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	apiKey, err = r.setEnabled(ctx, apiKey, plan.Enabled)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating API Key", err.Error())
		return
	}

	plan.Token = state.Token
	resp.Diagnostics.Append(flattenAPIKey(ctx, apiKey, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// setEnabled toggles the key when its enabled state differs from want, which
// is skipped when want is unknown (not configured).
func (r *apiKeyResource) setEnabled(ctx context.Context, apiKey *client.APIKey, want types.Bool) (*client.APIKey, error) {
	if want.IsUnknown() || want.IsNull() || want.ValueBool() == apiKey.Enabled {
		return apiKey, nil
	}
	if err := r.client.ToggleAPIKey(ctx, apiKey.ID); err != nil {
		return nil, fmt.Errorf("could not toggle API key ID %s: %w", apiKey.ID, err)
	}
	return r.client.GetAPIKey(ctx, apiKey.ID)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAPIKey(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error Deleting API Key",
			"Could not delete API key ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
