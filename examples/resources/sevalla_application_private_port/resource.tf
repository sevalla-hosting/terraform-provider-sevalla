resource "sevalla_application_private_port" "example" {
  application_id = sevalla_application.example.id
  process_id     = sevalla_application_process.web.id
  port           = 3000
  protocol       = "TCP"
}
