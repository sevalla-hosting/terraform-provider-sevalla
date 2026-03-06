data "sevalla_api_key" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "api_key_name" {
  value = data.sevalla_api_key.example.name
}
