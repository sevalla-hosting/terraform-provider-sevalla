data "sevalla_object_storage" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "endpoint" {
  value = data.sevalla_object_storage.example.endpoint
}
