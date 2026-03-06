package client

type Endpoint struct {
	Method string
	Path   string
}

var UsedEndpoints = []Endpoint{
	{Method: "GET", Path: "/v3/applications"},
	{Method: "POST", Path: "/v3/applications"},
	{Method: "GET", Path: "/v3/applications/{id}"},
	{Method: "PATCH", Path: "/v3/applications/{id}"},
	{Method: "DELETE", Path: "/v3/applications/{id}"},
	{Method: "POST", Path: "/v3/applications/{id}/activate"},
	{Method: "POST", Path: "/v3/applications/{id}/suspend"},
	{Method: "POST", Path: "/v3/applications/{id}/cdn/toggle"},
	{Method: "GET", Path: "/v3/resources/clusters"},
	{Method: "GET", Path: "/v3/resources/process-resource-types"},
	{Method: "GET", Path: "/v3/resources/database-resource-types"},

	{Method: "GET", Path: "/v3/databases"},
	{Method: "POST", Path: "/v3/databases"},
	{Method: "GET", Path: "/v3/databases/{id}"},
	{Method: "PATCH", Path: "/v3/databases/{id}"},
	{Method: "DELETE", Path: "/v3/databases/{id}"},
	{Method: "POST", Path: "/v3/databases/{id}/suspend"},
	{Method: "POST", Path: "/v3/databases/{id}/activate"},
	{Method: "POST", Path: "/v3/databases/{id}/external-connection/toggle"},
	{Method: "GET", Path: "/v3/databases/{id}/internal-connections"},
	{Method: "POST", Path: "/v3/databases/{id}/internal-connections"},
	{Method: "DELETE", Path: "/v3/databases/{id}/internal-connections/{conn_id}"},
	{Method: "GET", Path: "/v3/databases/{id}/ip-restriction"},
	{Method: "PUT", Path: "/v3/databases/{id}/ip-restriction"},

	{Method: "GET", Path: "/v3/static-sites"},
	{Method: "POST", Path: "/v3/static-sites"},
	{Method: "GET", Path: "/v3/static-sites/{id}"},
	{Method: "PATCH", Path: "/v3/static-sites/{id}"},
	{Method: "DELETE", Path: "/v3/static-sites/{id}"},

	{Method: "GET", Path: "/v3/load-balancers"},
	{Method: "POST", Path: "/v3/load-balancers"},
	{Method: "GET", Path: "/v3/load-balancers/{id}"},
	{Method: "PATCH", Path: "/v3/load-balancers/{id}"},
	{Method: "DELETE", Path: "/v3/load-balancers/{id}"},
	{Method: "GET", Path: "/v3/load-balancers/{id}/destinations"},
	{Method: "POST", Path: "/v3/load-balancers/{id}/destinations"},
	{Method: "DELETE", Path: "/v3/load-balancers/{id}/destinations/{dest_id}"},
	{Method: "POST", Path: "/v3/load-balancers/{id}/destinations/{dest_id}/toggle"},

	{Method: "GET", Path: "/v3/object-storage"},
	{Method: "POST", Path: "/v3/object-storage"},
	{Method: "GET", Path: "/v3/object-storage/{id}"},
	{Method: "PATCH", Path: "/v3/object-storage/{id}"},
	{Method: "DELETE", Path: "/v3/object-storage/{id}"},
	{Method: "POST", Path: "/v3/object-storage/{id}/domain"},
	{Method: "DELETE", Path: "/v3/object-storage/{id}/domain"},
	{Method: "GET", Path: "/v3/object-storage/{id}/cors"},
	{Method: "POST", Path: "/v3/object-storage/{id}/cors"},
	{Method: "PATCH", Path: "/v3/object-storage/{id}/cors/{policy_id}"},
	{Method: "DELETE", Path: "/v3/object-storage/{id}/cors/{policy_id}"},

	{Method: "GET", Path: "/v3/{service_type}/{id}/domains"},
	{Method: "GET", Path: "/v3/{service_type}/{id}/domains/{domain_id}"},
	{Method: "POST", Path: "/v3/{service_type}/{id}/domains"},
	{Method: "DELETE", Path: "/v3/{service_type}/{id}/domains/{domain_id}"},
	{Method: "POST", Path: "/v3/{service_type}/{id}/domains/{domain_id}/set-primary"},
	{Method: "POST", Path: "/v3/{service_type}/{id}/domains/{domain_id}/toggle"},
	{Method: "POST", Path: "/v3/{service_type}/{id}/domains/{domain_id}/refresh-status"},

	{Method: "GET", Path: "/v3/{service_type}/{id}/env-vars"},
	{Method: "POST", Path: "/v3/{service_type}/{id}/env-vars"},
	{Method: "PUT", Path: "/v3/{service_type}/{id}/env-vars/{env_var_id}"},
	{Method: "DELETE", Path: "/v3/{service_type}/{id}/env-vars/{env_var_id}"},

	// Pipelines
	{Method: "GET", Path: "/v3/pipelines"},
	{Method: "POST", Path: "/v3/pipelines"},
	{Method: "GET", Path: "/v3/pipelines/{id}"},
	{Method: "PATCH", Path: "/v3/pipelines/{id}"},
	{Method: "DELETE", Path: "/v3/pipelines/{id}"},
	{Method: "POST", Path: "/v3/pipelines/{id}/stages"},
	{Method: "DELETE", Path: "/v3/pipelines/{id}/stages/{stage_id}"},
	{Method: "POST", Path: "/v3/pipelines/{id}/stages/{stage_id}/apps"},
	{Method: "DELETE", Path: "/v3/pipelines/{id}/stages/{stage_id}/apps/{app_id}"},
	{Method: "POST", Path: "/v3/pipelines/{id}/preview/enable"},
	{Method: "POST", Path: "/v3/pipelines/{id}/preview/disable"},
	{Method: "PUT", Path: "/v3/pipelines/{id}/preview"},

	// Projects
	{Method: "GET", Path: "/v3/projects"},
	{Method: "POST", Path: "/v3/projects"},
	{Method: "GET", Path: "/v3/projects/{id}"},
	{Method: "PATCH", Path: "/v3/projects/{id}"},
	{Method: "DELETE", Path: "/v3/projects/{id}"},
	{Method: "POST", Path: "/v3/projects/{id}/services"},
	{Method: "DELETE", Path: "/v3/projects/{id}/services/{service_id}"},

	// Docker Registries
	{Method: "GET", Path: "/v3/docker-registries"},
	{Method: "POST", Path: "/v3/docker-registries"},
	{Method: "GET", Path: "/v3/docker-registries/{id}"},
	{Method: "PATCH", Path: "/v3/docker-registries/{id}"},
	{Method: "DELETE", Path: "/v3/docker-registries/{id}"},

	// Webhooks
	{Method: "GET", Path: "/v3/webhooks"},
	{Method: "POST", Path: "/v3/webhooks"},
	{Method: "GET", Path: "/v3/webhooks/{id}"},
	{Method: "PATCH", Path: "/v3/webhooks/{id}"},
	{Method: "DELETE", Path: "/v3/webhooks/{id}"},
	{Method: "POST", Path: "/v3/webhooks/{id}/toggle"},
	{Method: "POST", Path: "/v3/webhooks/{id}/roll-secret"},

	// API Keys
	{Method: "GET", Path: "/v3/api-keys"},
	{Method: "POST", Path: "/v3/api-keys"},
	{Method: "GET", Path: "/v3/api-keys/{id}"},
	{Method: "PATCH", Path: "/v3/api-keys/{id}"},
	{Method: "DELETE", Path: "/v3/api-keys/{id}"},
	{Method: "POST", Path: "/v3/api-keys/{id}/toggle"},
	{Method: "POST", Path: "/v3/api-keys/{id}/rotate"},

	// Global Environment Variables
	{Method: "GET", Path: "/v3/applications/global-env-vars"},
	{Method: "POST", Path: "/v3/applications/global-env-vars"},
	{Method: "PUT", Path: "/v3/applications/global-env-vars/{id}"},
	{Method: "DELETE", Path: "/v3/applications/global-env-vars/{id}"},

	// Users
	{Method: "GET", Path: "/v3/users"},
}
