package provider

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/client"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/apikey"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/application"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/database"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/dockerregistry"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/globalenvvar"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/loadbalancer"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/objectstorage"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/pipeline"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/project"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/referencedata"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/staticsite"
	"github.com/sevalla-hosting/terraform-provider-sevalla/internal/services/webhook"
)

var _ provider.Provider = &SevallaProvider{}

type SevallaProvider struct {
	version string
}

type SevallaProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SevallaProvider{
			version: version,
		}
	}
}

func (p *SevallaProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sevalla"
	resp.Version = p.version
}

func (p *SevallaProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Sevalla cloud infrastructure.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Sevalla API key. Can also be set via the SEVALLA_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *SevallaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SevallaProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("SEVALLA_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Sevalla API key must be set in the provider configuration or via the SEVALLA_API_KEY environment variable.",
		)
		return
	}

	opts := []client.Option{client.WithUserAgent("terraform-provider-sevalla/" + p.version)}
	if apiURL := os.Getenv("SEVALLA_API_URL"); apiURL != "" {
		parsed, err := url.Parse(apiURL)
		if err != nil || !strings.HasPrefix(parsed.Scheme, "https") {
			resp.Diagnostics.AddError(
				"Invalid API URL",
				"SEVALLA_API_URL must use the https:// scheme.",
			)
			return
		}
		opts = append(opts, client.WithBaseURL(apiURL))
	}
	c := client.NewClient(apiKey, opts...)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *SevallaProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Application
		application.NewResource,
		application.NewDomainResource,
		application.NewEnvironmentVariableResource,
		application.NewProcessResource,
		application.NewTCPProxyResource,
		application.NewPrivatePortResource,
		application.NewIPRestrictionResource,
		application.NewDeploymentHookResource,

		// Database
		database.NewResource,
		database.NewInternalConnectionResource,
		database.NewIPRestrictionResource,

		// Static Site
		staticsite.NewResource,
		staticsite.NewDomainResource,
		staticsite.NewEnvironmentVariableResource,

		// Load Balancer
		loadbalancer.NewResource,
		loadbalancer.NewDomainResource,
		loadbalancer.NewDestinationResource,

		// Object Storage
		objectstorage.NewResource,
		objectstorage.NewCORSPolicyResource,

		// Pipeline
		pipeline.NewResource,
		pipeline.NewStageResource,
		pipeline.NewStageApplicationResource,

		// Project
		project.NewResource,
		project.NewServiceResource,

		// Docker Registry
		dockerregistry.NewResource,

		// Webhook
		webhook.NewResource,

		// API Key
		apikey.NewResource,

		// Global Environment Variable
		globalenvvar.NewResource,
	}
}

func (p *SevallaProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Application
		application.NewDataSource,
		application.NewListDataSource,

		// Database
		database.NewDataSource,
		database.NewListDataSource,

		// Static Site
		staticsite.NewDataSource,
		staticsite.NewListDataSource,

		// Load Balancer
		loadbalancer.NewDataSource,
		loadbalancer.NewListDataSource,

		// Object Storage
		objectstorage.NewDataSource,

		// Pipeline
		pipeline.NewDataSource,

		// Project
		project.NewDataSource,

		// Docker Registry
		dockerregistry.NewDataSource,

		// Webhook
		webhook.NewDataSource,

		// API Key
		apikey.NewDataSource,

		// Reference Data
		referencedata.NewClustersDataSource,
		referencedata.NewProcessResourceTypesDataSource,
		referencedata.NewDatabaseResourceTypesDataSource,
		referencedata.NewUsersDataSource,
		referencedata.NewAPIKeyPermissionsDataSource,
		referencedata.NewAPIKeyRolesDataSource,
	}
}
