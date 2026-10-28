# Load balancer with round-robin or weighted routing
resource "sevalla_load_balancer" "example" {
  display_name = "my-load-balancer"
}

# Load balancer that routes by client location
resource "sevalla_load_balancer" "geo" {
  display_name = "my-geo-lb"
  type         = "GEO"
}
