# A new pipeline has a Dev (0) and a Production (1) stage.
# Insert a staging stage between them.
resource "sevalla_pipeline_stage" "staging" {
  pipeline_id   = sevalla_pipeline.example.id
  name          = "staging"
  insert_before = 1
}

# In a branch pipeline, each stage can deploy its own branch.
resource "sevalla_pipeline" "feature" {
  display_name = "feature-pipeline"
  type         = "branch"
}

resource "sevalla_pipeline_stage" "qa" {
  pipeline_id   = sevalla_pipeline.feature.id
  name          = "qa"
  insert_before = 1
  branch        = "develop"
}
