# Connect a database to an application
resource "sevalla_database_internal_connection" "example" {
  database_id    = sevalla_database.postgres.id
  application_id = sevalla_application.example.id
  target_type    = "app"
}
