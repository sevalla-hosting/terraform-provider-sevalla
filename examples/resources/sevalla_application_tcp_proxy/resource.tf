resource "sevalla_application_tcp_proxy" "example" {
  application_id = sevalla_application.example.id
  process_id     = sevalla_application_process.web.id
  port           = 5432
}
