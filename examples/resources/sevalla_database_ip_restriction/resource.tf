# Only allow external connections from an office network
resource "sevalla_database_ip_restriction" "example" {
  database_id = sevalla_database.postgres.id
  type        = "allow"
  is_enabled  = true
  ip_list     = ["203.0.113.0/24"]
}
