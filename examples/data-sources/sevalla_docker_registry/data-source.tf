data "sevalla_docker_registry" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "registry_name" {
  value = data.sevalla_docker_registry.example.name
}
