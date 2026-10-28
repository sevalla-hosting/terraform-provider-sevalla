# Restart the application with the last build whenever an environment
# variable changes.
resource "sevalla_application_deployment" "restart" {
  application_id = sevalla_application.example.id
  is_restart     = true

  triggers = {
    api_url = sevalla_application_environment_variable.api_url.value
  }
}

# Build and deploy a specific commit, and fail the apply if the deployment fails.
resource "sevalla_application_deployment" "release" {
  application_id      = sevalla_application.example.id
  is_restart          = false
  branch              = "main"
  commit_sha          = "0123456789abcdef0123456789abcdef01234567"
  wait_for_completion = true
}
