package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
)

var (
	_ resource.Resource                = &applicationDomainResource{}
	_ resource.ResourceWithConfigure   = &applicationDomainResource{}
	_ resource.ResourceWithImportState = &applicationDomainResource{}
)

type applicationDomainResource struct {
	client *client.SevallaClient
}

type ApplicationDomainResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	IsPrimary     types.Bool   `tfsdk:"is_primary"`
	IsWildcard    types.Bool   `tfsdk:"is_wildcard"`
	IsEnabled     types.Bool   `tfsdk:"is_enabled"`
	Status        types.String `tfsdk:"status"`
	CustomSSLCert types.String `tfsdk:"custom_ssl_cert"`
	CustomSSLKey  types.String `tfsdk:"custom_ssl_key"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewDomainResource() resource.Resource {
	return &applicationDomainResource{}
}

func (r *applicationDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_domain"
}

func (r *applicationDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a domain for a Sevalla application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this domain belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The domain name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The domain type (custom or system).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_primary": schema.BoolAttribute{
				Description: "Whether this is the primary domain.",
				Computed:    true,
			},
			"is_wildcard": schema.BoolAttribute{
				Description: "Whether this is a wildcard domain.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the domain is enabled.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The status of the domain.",
				Computed:    true,
			},
			"custom_ssl_cert": schema.StringAttribute{
				Description: "The custom SSL certificate in PEM format.",
				Optional:    true,
				Sensitive:   true,
			},
			"custom_ssl_key": schema.StringAttribute{
				Description: "The custom SSL private key in PEM format.",
				Optional:    true,
				Sensitive:   true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the domain was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp when the domain was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *applicationDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	createReq := &client.CreateDomainRequest{
		DomainName: plan.Name.ValueString(),
	}
	if !plan.IsWildcard.IsNull() && !plan.IsWildcard.IsUnknown() {
		v := plan.IsWildcard.ValueBool()
		createReq.IsWildcard = &v
	}
	if !plan.CustomSSLCert.IsNull() && !plan.CustomSSLCert.IsUnknown() {
		v := plan.CustomSSLCert.ValueString()
		createReq.CustomSSLCert = &v
	}
	if !plan.CustomSSLKey.IsNull() && !plan.CustomSSLKey.IsUnknown() {
		v := plan.CustomSSLKey.ValueString()
		createReq.CustomSSLKey = &v
	}
	domain, err := r.client.CreateDomain(ctx, "/applications", appID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application Domain", err.Error())
		return
	}

	flattenDomain(domain, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	domain, err := r.client.GetDomain(ctx, "/applications", appID, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application Domain", err.Error())
		return
	}

	flattenDomain(domain, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationDomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Application domains cannot be updated. Delete and recreate the domain instead.",
	)
}

func (r *applicationDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(ctx, "/applications", state.ApplicationID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Application Domain", err.Error())
		return
	}
}

func (r *applicationDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'application_id/domain_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenDomain(domain *client.Domain, model *ApplicationDomainResourceModel) {
	model.ID = types.StringValue(domain.ID)
	model.Name = types.StringValue(domain.Name)
	model.Type = types.StringValue(domain.Type)
	model.IsPrimary = types.BoolValue(domain.IsPrimary)
	model.IsWildcard = types.BoolValue(domain.IsWildcard)
	model.IsEnabled = types.BoolValue(domain.IsEnabled)
	model.Status = optionalString(domain.Status)
	model.CreatedAt = types.StringValue(domain.CreatedAt)
	model.UpdatedAt = types.StringValue(domain.UpdatedAt)
}
