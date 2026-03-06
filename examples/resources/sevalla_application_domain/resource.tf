# Standard custom domain
resource "sevalla_application_domain" "example" {
  application_id = sevalla_application.example.id
  name           = "app.example.com"
}

# Wildcard domain
resource "sevalla_application_domain" "wildcard" {
  application_id = sevalla_application.example.id
  name           = "*.example.com"
  is_wildcard    = true
}

# Domain with custom SSL certificate
resource "sevalla_application_domain" "custom_ssl" {
  application_id = sevalla_application.example.id
  name           = "secure.example.com"
  custom_ssl_cert = file("certs/cert.pem")
  custom_ssl_key  = file("certs/key.pem")
}
