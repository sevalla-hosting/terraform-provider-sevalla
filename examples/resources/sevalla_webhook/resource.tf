resource "sevalla_webhook" "example" {
  endpoint       = "https://hooks.example.com/sevalla"
  description    = "Notify on deployments"
  allowed_events = ["deployment.started", "deployment.completed", "deployment.failed"]
}
