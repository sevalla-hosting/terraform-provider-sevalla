# Application destination
resource "sevalla_load_balancer_destination" "app" {
  load_balancer_id = sevalla_load_balancer.example.id
  service_id       = sevalla_application.example.id
  service_type     = "APP"
}

# External URL destination
resource "sevalla_load_balancer_destination" "external" {
  load_balancer_id = sevalla_load_balancer.example.id
  service_type     = "EXTERNAL"
  service_id       = "external"
  url              = "https://legacy.example.com"
  weight           = 20
}
