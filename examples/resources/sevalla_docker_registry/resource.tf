# Docker Hub registry
resource "sevalla_docker_registry" "dockerhub" {
  name     = "my-dockerhub"
  registry = "dockerHub"
  username = var.dockerhub_username
  secret   = var.dockerhub_token
}

# GitHub Container Registry
resource "sevalla_docker_registry" "ghcr" {
  name     = "my-ghcr"
  registry = "github"
  username = var.github_username
  secret   = var.github_token
}
