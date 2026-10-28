package loadbalancer

import (
	"context"
	"fmt"
	"strings"

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
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/writeonly"
)

var (
	_ resource.Resource                = &loadBalancerDomainResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerDomainResource{}
	_ resource.ResourceWithImportState = &loadBalancerDomainResource{}
)

type loadBalancerDomainResource struct {
	client *client.SevallaClient
}

type LoadBalancerDomainResourceModel struct {
	ID             types.String `tfsdk:"id"`
	LoadBalancerID types.String `tfsdk:"load_balancer_id"`
	Name           types.String `tfsdk:"name"`
	IsWildcard     types.Bool   `tfsdk:"is_wildcard"`
	CustomSSLCert  types.String `tfsdk:"custom_ssl_cert"`
	CustomSSLKey   types.String `tfsdk:"custom_ssl_key"`
	Type           types.String `tfsdk:"type"`
	IsPrimary      types.Bool   `tfsdk:"is_primary"`
	IsEnabled      types.Bool   `tfsdk:"is_enabled"`
	Status         types.String `tfsdk:"status"`
	DNSRecords     types.String `tfsdk:"dns_records"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func NewDomainResource() resource.Resource {
	return &loadBalancerDomainResource{}
}

func (r *loadBalancerDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_domain"
}

func (r *loadBalancerDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a domain for a Sevalla load balancer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"load_balancer_id": schema.StringAttribute{
				Description: "The ID of the load balancer this domain belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The domain name.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_wildcard": schema.BoolAttribute{
				Description: "Whether the domain is a wildcard domain that matches all subdomains. Defaults to `false`. Changing this creates a new domain.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"custom_ssl_cert": schema.StringAttribute{
				Description: "Custom SSL certificate in PEM format, used instead of an auto-provisioned one. The API does not return it, so it is not checked for drift. Changing it creates a new domain, except after an import, when the configured value is only recorded.",
				Optional:    true,
				Sensitive:   true,
				Validators:  []validator.String{stringvalidator.AlsoRequires(path.MatchRoot("custom_ssl_key"))},
				PlanModifiers: []planmodifier.String{
					writeonly.RequiresReplace(),
				},
			},
			"custom_ssl_key": schema.StringAttribute{
				Description: "Private key of the custom SSL certificate in PEM format. Required with `custom_ssl_cert`. Changing it creates a new domain, except after an import, when the configured value is only recorded.",
				Optional:    true,
				Sensitive:   true,
				Validators:  []validator.String{stringvalidator.AlsoRequires(path.MatchRoot("custom_ssl_cert"))},
				PlanModifiers: []planmodifier.String{
					writeonly.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The domain type: `CUSTOM` for user-added domains or `SYSTEM` for platform-assigned ones.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_primary": schema.BoolAttribute{
				Description: "Whether this is the primary domain.",
				Computed:    true,
			},
			"is_enabled": schema.BoolAttribute{
				Description: "Whether the domain is enabled.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The verification status: `VERIFICATION`, `POINTING`, `ACTIVE` or `TIMED_OUT`.",
				Computed:    true,
			},
			"dns_records": schema.StringAttribute{
				Description: "JSON-encoded DNS records to create for the domain (e.g. `[{\"type\":\"CNAME\",\"name\":\"app.example.com\",\"value\":\"...\"}]`), or null when none are needed. Decode it with `jsondecode()`.",
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

func (r *loadBalancerDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loadBalancerDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadBalancerDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbID := plan.LoadBalancerID.ValueString()
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
	domain, err := r.client.CreateDomain(ctx, "/load-balancers", lbID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Load Balancer Domain", err.Error())
		return
	}

	flattenLBDomain(domain, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *loadBalancerDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadBalancerDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbID := state.LoadBalancerID.ValueString()
	domain, err := r.client.GetDomain(ctx, "/load-balancers", lbID, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Load Balancer Domain", err.Error())
		return
	}

	flattenLBDomain(domain, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is only reached when custom_ssl_cert or custom_ssl_key is first set
// after an import. The API never returns them, so they are only recorded.
func (r *loadBalancerDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadBalancerDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.CustomSSLCert = plan.CustomSSLCert
	state.CustomSSLKey = plan.CustomSSLKey
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *loadBalancerDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadBalancerDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(ctx, "/load-balancers", state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error Deleting Load Balancer Domain", err.Error())
		return
	}
}

func (r *loadBalancerDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'load_balancer_id/domain_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("load_balancer_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	resp.Diagnostics.Append(writeonly.MarkImported(ctx, resp.Private)...)
}

func flattenLBDomain(domain *client.Domain, model *LoadBalancerDomainResourceModel) {
	model.ID = types.StringValue(domain.ID)
	// Keep the configured spelling when the API only changes the case.
	if !strings.EqualFold(model.Name.ValueString(), domain.Name) {
		model.Name = types.StringValue(domain.Name)
	}
	model.IsWildcard = types.BoolValue(domain.IsWildcard)
	model.Type = types.StringValue(domain.Type)
	model.IsPrimary = types.BoolValue(domain.IsPrimary)
	model.IsEnabled = types.BoolValue(domain.IsEnabled)
	if domain.Status != nil {
		model.Status = types.StringValue(*domain.Status)
	} else {
		model.Status = types.StringNull()
	}
	model.CreatedAt = types.StringValue(domain.CreatedAt)
	model.UpdatedAt = types.StringValue(domain.UpdatedAt)
	model.DNSRecords = types.StringNull()
	if domain.DnsRecords != nil && string(*domain.DnsRecords) != "null" {
		model.DNSRecords = types.StringValue(string(*domain.DnsRecords))
	}
}
