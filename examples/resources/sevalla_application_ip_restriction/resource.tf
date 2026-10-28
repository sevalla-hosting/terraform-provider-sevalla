# Only allow traffic from an office network and a VPN server
resource "sevalla_application_ip_restriction" "example" {
  application_id = sevalla_application.example.id
  type           = "allow"
  is_enabled     = true
  ip_list        = ["203.0.113.0/24", "198.51.100.50"]
}
