# Attach an application to a project
resource "sevalla_project_service" "app" {
  project_id   = sevalla_project.example.id
  service_id   = sevalla_application.example.id
  service_type = "app"
}

# Attach a database to a project
resource "sevalla_project_service" "db" {
  project_id   = sevalla_project.example.id
  service_id   = sevalla_database.postgres.id
  service_type = "database"
}
