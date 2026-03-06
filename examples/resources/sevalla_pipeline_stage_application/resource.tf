resource "sevalla_pipeline_stage_application" "example" {
  pipeline_id    = sevalla_pipeline.example.id
  stage_id       = sevalla_pipeline_stage.staging.id
  application_id = sevalla_application.example.id
}
