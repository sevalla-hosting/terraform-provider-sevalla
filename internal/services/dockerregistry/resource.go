package dockerregistry

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
	_ resource.Resource                = &dockerRegistryResource{}
	_ resource.ResourceWithConfigure   = &dockerRegistryResource{}
	_ resource.ResourceWithImportState = &dockerRegistryResource{}
)

type dockerRegistryResource struct {
	client *client.SevallaClient
}

func NewResource() resource.Resource {
	return &dockerRegistryResource{}
}

func (r *dockerRegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_registry"
}

func (r *dockerRegistryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Sevalla Docker registry credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Docker registry credential.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Docker registry credential.",
				Required:    true,
			},
			"registry": schema.StringAttribute{
				Description: "The Docker registry type (e.g., gcr, ecr, dockerHub, github, gitlab, digitalOcean, custom).",
				Optional:    true,
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for the Docker registry.",
				Required:    true,
				Sensitive:   true,
			},
			"secret": schema.StringAttribute{
				Description: "The secret (password/token) for the Docker registry. Write-only; not returned by the API.",
				Required:    true,
				Sensitive:   true,
			},
			"company_id": schema.StringAttribute{
				Description: "The company ID that owns the Docker registry credential.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the Docker registry credential was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the Docker registry credential was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *dockerRegistryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dockerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DockerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildCreateRequest(&plan)

	registry, err := r.client.CreateDockerRegistry(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Docker Registry",
			"Could not create Docker registry, unexpected error: "+err.Error(),
		)
		return
	}

	// Preserve the secret from plan since it is write-only.
	secret := plan.Secret
	flattenDockerRegistry(registry, &plan)
	plan.Secret = secret

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dockerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DockerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.GetDockerRegistry(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Docker Registry",
			"Could not read Docker registry ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Preserve the secret from state since it is write-only.
	secret := state.Secret
	flattenDockerRegistry(registry, &state)
	state.Secret = secret

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dockerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DockerRegistryResourceModel
	var state DockerRegistryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildUpdateRequest(&plan, &state)

	registry, err := r.client.UpdateDockerRegistry(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Docker Registry",
			"Could not update Docker registry ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Preserve the secret from plan since it is write-only.
	secret := plan.Secret
	flattenDockerRegistry(registry, &plan)
	plan.Secret = secret

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dockerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DockerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDockerRegistry(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Docker Registry",
			"Could not delete Docker registry ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *dockerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
