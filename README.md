<div align="center">

# Sevalla Terraform Provider

**Manage your Sevalla infrastructure as code.**

[![Tests](https://github.com/sevalla-hosting/terraform-provider-sevalla/actions/workflows/test.yml/badge.svg)](https://github.com/sevalla-hosting/terraform-provider-sevalla/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Terraform](https://img.shields.io/badge/Terraform-1.0+-844FBA?logo=terraform&logoColor=white)](https://www.terraform.io/)
[![Registry](https://img.shields.io/badge/Registry-sevalla--hosting/sevalla-844FBA?logo=terraform&logoColor=white)](https://registry.terraform.io/providers/sevalla-hosting/sevalla/latest)

</div>

---

The official [Terraform](https://www.terraform.io/) provider for the [Sevalla](https://sevalla.com) cloud platform. Deploy applications, provision databases, manage static sites, configure load balancers, and more — all defined in Terraform configuration.

## Quick Start

```hcl
terraform {
  required_providers {
    sevalla = {
      source  = "sevalla-hosting/sevalla"
      version = "~> 0.1"
    }
  }
}

provider "sevalla" {
  # Set via SEVALLA_API_KEY environment variable
}

data "sevalla_clusters" "all" {}

resource "sevalla_application" "web" {
  display_name = "my-web-app"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/example/app"
}
```

```bash
export SEVALLA_API_KEY="your-api-key"
terraform init
terraform plan
terraform apply
```

## Authentication

Set your API key via the `SEVALLA_API_KEY` environment variable or in the provider block:

```hcl
provider "sevalla" {
  api_key = var.sevalla_api_key
}
```

Generate an API key from **Settings > API Keys** in the [Sevalla dashboard](https://app.sevalla.com/api-keys).

## Resources

| Resource | Description |
|----------|-------------|
| `sevalla_application` | Web application |
| `sevalla_application_domain` | Custom domain for an application |
| `sevalla_application_env_var` | Environment variable for an application |
| `sevalla_application_process` | Process (web, worker, cron) for an application |
| `sevalla_application_tcp_proxy` | TCP proxy for an application |
| `sevalla_application_private_port` | Private port for an application |
| `sevalla_application_ip_restriction` | IP restriction rules for an application |
| `sevalla_application_deployment_hook` | Deployment hook for an application |
| `sevalla_database` | Managed database (PostgreSQL, MySQL, Redis) |
| `sevalla_database_internal_connection` | Internal connection between database and application |
| `sevalla_database_ip_restriction` | IP restriction rules for a database |
| `sevalla_static_site` | Static site |
| `sevalla_static_site_domain` | Custom domain for a static site |
| `sevalla_static_site_env_var` | Environment variable for a static site |
| `sevalla_load_balancer` | Load balancer |
| `sevalla_load_balancer_domain` | Custom domain for a load balancer |
| `sevalla_load_balancer_destination` | Destination application for a load balancer |
| `sevalla_object_storage` | Object storage bucket |
| `sevalla_object_storage_cors_policy` | CORS policy for an object storage bucket |
| `sevalla_pipeline` | Deployment pipeline |
| `sevalla_pipeline_stage` | Stage in a pipeline |
| `sevalla_pipeline_stage_application` | Application in a pipeline stage |
| `sevalla_project` | Project |
| `sevalla_project_service` | Service in a project |
| `sevalla_docker_registry` | Docker registry credentials |
| `sevalla_webhook` | Webhook |
| `sevalla_api_key` | API key |
| `sevalla_global_env_var` | Global environment variable |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `sevalla_application` | Look up an application |
| `sevalla_applications` | List all applications |
| `sevalla_database` | Look up a database |
| `sevalla_databases` | List all databases |
| `sevalla_static_site` | Look up a static site |
| `sevalla_static_sites` | List all static sites |
| `sevalla_load_balancer` | Look up a load balancer |
| `sevalla_load_balancers` | List all load balancers |
| `sevalla_object_storage` | Look up an object storage bucket |
| `sevalla_pipeline` | Look up a pipeline |
| `sevalla_project` | Look up a project |
| `sevalla_docker_registry` | Look up a Docker registry |
| `sevalla_webhook` | Look up a webhook |
| `sevalla_api_key` | Look up an API key |
| `sevalla_clusters` | List available clusters |
| `sevalla_process_resource_types` | List process resource types |
| `sevalla_database_resource_types` | List database resource types |
| `sevalla_users` | List users |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SEVALLA_API_KEY` | API key for authentication |
| `SEVALLA_API_URL` | Override API base URL (must be HTTPS) |

## Development

**Requirements:** Go 1.24+

```bash
git clone https://github.com/sevalla-hosting/terraform-provider-sevalla.git
cd terraform-provider-sevalla
```

```bash
make build      # Compile the provider binary
make test       # Run unit tests
make testacc    # Run acceptance tests (requires SEVALLA_API_KEY)
```

To use a locally built provider, add a dev override to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "sevalla-hosting/sevalla" = "/path/to/terraform-provider-sevalla"
  }
  direct {}
}
```

## License

[MIT](LICENSE)
