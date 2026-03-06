# Standard domain
resource "sevalla_static_site_domain" "example" {
  static_site_id = sevalla_static_site.example.id
  name           = "docs.example.com"
}

# Wildcard domain with custom SSL
resource "sevalla_static_site_domain" "wildcard" {
  static_site_id = sevalla_static_site.example.id
  name           = "example.com"
  is_wildcard    = true
  custom_ssl_cert = file("certs/example.com.pem")
  custom_ssl_key  = file("certs/example.com.key")
}
