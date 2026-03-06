data "sevalla_databases" "all" {}

output "database_names" {
  value = [for db in data.sevalla_databases.all.databases : db.display_name]
}
