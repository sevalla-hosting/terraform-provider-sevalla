data "sevalla_static_sites" "all" {}

output "static_site_names" {
  value = [for s in data.sevalla_static_sites.all.static_sites : s.display_name]
}
