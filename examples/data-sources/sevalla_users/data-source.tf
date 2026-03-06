data "sevalla_users" "all" {}

output "user_emails" {
  value = [for u in data.sevalla_users.all.users : u.email]
}
