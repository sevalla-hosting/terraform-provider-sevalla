resource "sevalla_database_ip_restriction" "example" {
  database_id = sevalla_database.postgres.id
  type        = "allow"
  enabled     = true

  rules {
    address     = "203.0.113.0/24"
    description = "Office network"
  }
}
