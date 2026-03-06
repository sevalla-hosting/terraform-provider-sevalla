data "sevalla_process_resource_types" "all" {}

output "resource_type_names" {
  value = [for rt in data.sevalla_process_resource_types.all.process_resource_types : rt.name]
}
