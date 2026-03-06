package staticsite

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
	_ resource.Resource                = &staticSiteDomainResource{}
	_ resource.ResourceWithConfigure   = &staticSiteDomainResource{}
	_ resource.ResourceWithImportState = &staticSiteDomainResource{}
)

type staticSiteDomainResource struct {
	client *client.SevallaClient
}

type StaticSiteDomainResourceModel struct {
	ID            types.String `tfsdk:"id"`
	StaticSiteID  types.String `tfsdk:"static_site_id"`
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
	return &staticSiteDomainResource{}
}

func (r *staticSiteDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_site_domain"
}

func (r *staticSiteDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a domain for a Sevalla static site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"static_site_id": schema.StringAttribute{
				Description: "The ID of the static site this domain belongs to.",
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
				Description: "Whether this should be a wildcard domain that matches all subdomains.",
				Optional:    true,
				Computed:    true,
			},
			"custom_ssl_cert": schema.StringAttribute{
				Description: "Custom SSL certificate in PEM format. When provided, the platform uses this certificate instead of auto-provisioning one.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_ssl_key": schema.StringAttribute{
				Description: "Private key for the custom SSL certificate in PEM format. Required when custom_ssl_cert is provided.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *staticSiteDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *staticSiteDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := plan.StaticSiteID.ValueString()
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
	domain, err := r.client.CreateDomain(ctx, "/static-sites", siteID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Static Site Domain", err.Error())
		return
	}

	flattenStaticSiteDomain(domain, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *staticSiteDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.StaticSiteID.ValueString()
	domain, err := r.client.GetDomain(ctx, "/static-sites", siteID, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Static Site Domain", err.Error())
		return
	}

	flattenStaticSiteDomain(domain, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *staticSiteDomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Static site domains cannot be updated. Delete and recreate the domain instead.",
	)
}

func (r *staticSiteDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StaticSiteDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(ctx, "/static-sites", state.StaticSiteID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Static Site Domain", err.Error())
		return
	}
}

func (r *staticSiteDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'static_site_id/domain_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("static_site_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func flattenStaticSiteDomain(domain *client.Domain, model *StaticSiteDomainResourceModel) {
	model.ID = types.StringValue(domain.ID)
	model.Name = types.StringValue(domain.Name)
	model.Type = types.StringValue(domain.Type)
	model.IsPrimary = types.BoolValue(domain.IsPrimary)
	model.IsWildcard = types.BoolValue(domain.IsWildcard)
	model.IsEnabled = types.BoolValue(domain.IsEnabled)
	if domain.Status != nil {
		model.Status = types.StringValue(*domain.Status)
	} else {
		model.Status = types.StringNull()
	}
	model.CreatedAt = types.StringValue(domain.CreatedAt)
	model.UpdatedAt = types.StringValue(domain.UpdatedAt)
}
