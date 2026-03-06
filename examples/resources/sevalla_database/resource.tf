# PostgreSQL database with extensions
resource "sevalla_database" "postgres" {
  display_name     = "my-postgres-db"
  type             = "postgresql"
  version          = "16"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "myapp"
  db_user          = "myapp_user"
  db_password      = var.db_password

  extensions {
    enable_postgis = true
    enable_vector  = true
  }
}

# Redis database
resource "sevalla_database" "redis" {
  display_name     = "my-redis-cache"
  type             = "redis"
  version          = "7"
  cluster_id       = data.sevalla_clusters.all.clusters[0].id
  resource_type_id = data.sevalla_database_resource_types.all.database_resource_types[0].id
  db_name          = "cache"
  db_password      = var.redis_password
}
