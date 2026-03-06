data "sevalla_database" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "db_internal_host" {
  value = data.sevalla_database.example.internal_hostname
}
