data "sevalla_api_key_roles" "all" {}

# API key with a predefined role
resource "sevalla_api_key" "ci" {
  name       = "ci-cd-key"
  expires_at = "2027-01-01T00:00:00Z"
  role_ids   = [data.sevalla_api_key_roles.all.roles[0].id]
}

# Read-only API key scoped with capabilities
resource "sevalla_api_key" "monitoring" {
  name = "monitoring-key"

  capabilities = [
    { permission = "APP:READ" },
    { permission = "DATABASE:READ" },
  ]
}
