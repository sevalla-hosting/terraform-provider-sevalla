# Production-only variable
resource "sevalla_static_site_environment_variable" "api_url" {
  static_site_id = sevalla_static_site.example.id
  key            = "API_BASE_URL"
  value          = "https://api.example.com"
  is_production  = true
  is_preview     = false
}

# Branch-scoped preview variable
resource "sevalla_static_site_environment_variable" "preview" {
  static_site_id = sevalla_static_site.example.id
  key            = "API_BASE_URL"
  value          = "https://staging-api.example.com"
  is_production  = false
  is_preview     = true
  branch         = "develop"
}
