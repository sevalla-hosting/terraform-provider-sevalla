resource "sevalla_object_storage_cors_policy" "example" {
  object_storage_id = sevalla_object_storage.example.id
  allowed_origins   = ["https://example.com", "https://app.example.com"]
  allowed_methods   = ["GET", "PUT", "POST"]
  allowed_headers   = ["Content-Type", "Authorization"]
  max_age           = 3600
}
