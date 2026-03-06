data "sevalla_application" "example" {
  id = "fb5e5168-4281-4bec-94c5-0d1584e9e657"
}

output "app_name" {
  value = data.sevalla_application.example.display_name
}
