resource "sevalla_global_environment_variable" "example" {
  key          = "LOG_LEVEL"
  value        = "info"
  is_runtime   = true
  is_buildtime = false
}
