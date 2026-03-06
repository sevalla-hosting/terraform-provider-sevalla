data "sevalla_webhook" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "webhook_endpoint" {
  value = data.sevalla_webhook.example.endpoint
}
