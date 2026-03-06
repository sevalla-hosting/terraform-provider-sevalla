data "sevalla_load_balancers" "all" {}

output "load_balancer_names" {
  value = [for lb in data.sevalla_load_balancers.all.load_balancers : lb.display_name]
}
