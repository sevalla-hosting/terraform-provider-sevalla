# Standard domain
resource "sevalla_load_balancer_domain" "example" {
  load_balancer_id = sevalla_load_balancer.example.id
  name             = "lb.example.com"
}

# Wildcard domain with custom SSL
resource "sevalla_load_balancer_domain" "wildcard" {
  load_balancer_id = sevalla_load_balancer.example.id
  name             = "example.com"
  is_wildcard      = true
  custom_ssl_cert  = file("certs/example.com.pem")
  custom_ssl_key   = file("certs/example.com.key")
}
