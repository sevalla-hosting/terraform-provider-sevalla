# Standard load balancer
resource "sevalla_load_balancer" "example" {
  display_name = "my-load-balancer"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  port         = 443
  protocol     = "HTTPS"
  algorithm    = "round-robin"
}

# Geo-routed load balancer
resource "sevalla_load_balancer" "geo" {
  display_name = "my-geo-lb"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  port         = 443
  protocol     = "HTTPS"
  algorithm    = "round-robin"
  type         = "GEO"
}
