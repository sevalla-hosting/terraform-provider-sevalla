# Staging stage
resource "sevalla_pipeline_stage" "staging" {
  pipeline_id = sevalla_pipeline.example.id
  name        = "staging"
  branch      = "develop"
}

# Production stage inserted after staging
resource "sevalla_pipeline_stage" "production" {
  pipeline_id   = sevalla_pipeline.example.id
  name          = "production"
  insert_before = 2
}
