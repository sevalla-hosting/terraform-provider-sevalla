resource "sevalla_application_ip_restriction" "example" {
  application_id = sevalla_application.example.id
  type           = "allow"
  enabled        = true

  rules {
    address     = "203.0.113.0/24"
    description = "Office network"
  }

  rules {
    address     = "198.51.100.50"
    description = "VPN server"
  }
}
