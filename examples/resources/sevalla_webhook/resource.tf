resource "sevalla_webhook" "example" {
  endpoint       = "https://hooks.example.com/sevalla"
  description    = "Notify on deployments"
  allowed_events = ["APP_DEPLOY", "STATIC_SITE_DEPLOY"]
}
