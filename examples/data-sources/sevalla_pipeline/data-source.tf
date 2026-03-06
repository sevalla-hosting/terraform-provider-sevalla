data "sevalla_pipeline" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "pipeline_type" {
  value = data.sevalla_pipeline.example.type
}
