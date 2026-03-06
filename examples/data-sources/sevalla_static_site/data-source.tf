data "sevalla_static_site" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "site_hostname" {
  value = data.sevalla_static_site.example.hostname
}
