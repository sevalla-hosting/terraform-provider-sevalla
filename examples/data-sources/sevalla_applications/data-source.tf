data "sevalla_applications" "all" {}

output "application_names" {
  value = [for app in data.sevalla_applications.all.applications : app.display_name]
}
