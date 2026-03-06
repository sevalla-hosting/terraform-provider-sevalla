data "sevalla_clusters" "all" {}

output "cluster_locations" {
  value = [for c in data.sevalla_clusters.all.clusters : c.display_name]
}
