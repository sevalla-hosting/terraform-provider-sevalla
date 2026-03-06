# Available at both runtime and build time (defaults)
resource "sevalla_application_environment_variable" "example" {
  application_id = sevalla_application.example.id
  key            = "DATABASE_URL"
  value          = "postgres://user:pass@host:5432/db"
}

# Build-time only variable
resource "sevalla_application_environment_variable" "build_only" {
  application_id = sevalla_application.example.id
  key            = "NPM_TOKEN"
  value          = var.npm_token
  is_runtime     = false
  is_buildtime   = true
}
