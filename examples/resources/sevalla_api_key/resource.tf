# Full-access API key
resource "sevalla_api_key" "example" {
  name       = "ci-cd-key"
  expires_at = "2027-01-01T00:00:00Z"
}

# Scoped API key with specific capabilities
resource "sevalla_api_key" "readonly" {
  name         = "monitoring-key"
  expires_at   = "2027-06-01T00:00:00Z"
  capabilities = ["APP:READ", "DATABASE:READ"]
}
