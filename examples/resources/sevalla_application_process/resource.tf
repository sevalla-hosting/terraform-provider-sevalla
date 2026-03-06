# Web process with manual scaling
resource "sevalla_application_process" "web" {
  application_id = sevalla_application.example.id
  display_name   = "web"
  type           = "web"
  entrypoint     = "npm start"
  port           = 3000
  resource_type_id = data.sevalla_process_resource_types.all.process_resource_types[0].id

  scaling_strategy {
    type           = "manual"
    instance_count = 2
  }
}

# Web process with horizontal autoscaling and health probes
resource "sevalla_application_process" "web_autoscale" {
  application_id = sevalla_application.example.id
  display_name   = "web-autoscale"
  type           = "web"
  entrypoint     = "npm start"
  port           = 8080
  resource_type_id = data.sevalla_process_resource_types.all.process_resource_types[0].id

  scaling_strategy {
    type                = "horizontal"
    min_instance_count  = 1
    max_instance_count  = 5
    target_cpu_percent  = 80
  }

  liveness_probe {
    http_get {
      path = "/healthz"
      port = 8080
    }
    initial_delay_seconds = 10
    period_seconds        = 30
    failure_threshold     = 3
  }

  readiness_probe {
    http_get {
      path = "/ready"
      port = 8080
    }
    period_seconds    = 10
    failure_threshold = 3
  }
}

# Worker process
resource "sevalla_application_process" "worker" {
  application_id = sevalla_application.example.id
  display_name   = "worker"
  type           = "worker"
  entrypoint     = "npm run worker"
  resource_type_id = data.sevalla_process_resource_types.all.process_resource_types[0].id

  scaling_strategy {
    type           = "manual"
    instance_count = 1
  }
}

# Cron process with timezone
resource "sevalla_application_process" "cron" {
  application_id = sevalla_application.example.id
  display_name   = "cleanup"
  type           = "cron"
  entrypoint     = "npm run cleanup"
  resource_type_id = data.sevalla_process_resource_types.all.process_resource_types[0].id
  schedule       = "0 */6 * * *"
  time_zone      = "America/New_York"

  scaling_strategy {
    type           = "manual"
    instance_count = 1
  }
}

# Job process that runs before deployment
resource "sevalla_application_process" "migrate" {
  application_id = sevalla_application.example.id
  display_name   = "migrate"
  type           = "job"
  entrypoint     = "npm run db:migrate"
  resource_type_id = data.sevalla_process_resource_types.all.process_resource_types[0].id
  job_start_policy = "beforeDeployment"

  scaling_strategy {
    type           = "manual"
    instance_count = 1
  }
}
