data "sevalla_load_balancer" "example" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}

output "lb_type" {
  value = data.sevalla_load_balancer.example.type
}
