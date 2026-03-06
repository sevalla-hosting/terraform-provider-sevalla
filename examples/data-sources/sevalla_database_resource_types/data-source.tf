data "sevalla_database_resource_types" "all" {}

output "database_resource_type_names" {
  value = [for rt in data.sevalla_database_resource_types.all.database_resource_types : rt.name]
}
