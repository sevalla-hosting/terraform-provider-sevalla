# Deploy from a public Git repository
resource "sevalla_application" "example" {
  display_name = "my-web-app"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "publicGit"
  repo_url     = "https://github.com/example/app"
}

# Deploy from a private GitHub repository
resource "sevalla_application" "private" {
  display_name   = "my-private-app"
  cluster_id     = data.sevalla_clusters.all.clusters[0].id
  source         = "privateGit"
  git_type       = "github"
  repo_url       = "https://github.com/myorg/private-repo"
  default_branch = "main"
  auto_deploy    = true
  build_type     = "nixpacks"
}

# Deploy from a Docker image
resource "sevalla_application" "docker" {
  display_name = "my-docker-app"
  cluster_id   = data.sevalla_clusters.all.clusters[0].id
  source       = "dockerImage"
  docker_image = "nginx:latest"
}
