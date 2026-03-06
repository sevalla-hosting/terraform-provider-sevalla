# Static site from a public repository
resource "sevalla_static_site" "example" {
  display_name   = "my-docs-site"
  repo_url       = "https://github.com/example/docs"
  default_branch = "main"
  build_command  = "npm run build"
}

# Static site from a private GitHub repository
resource "sevalla_static_site" "private" {
  display_name        = "my-private-site"
  source              = "privateGit"
  git_type            = "github"
  repo_url            = "https://github.com/myorg/website"
  default_branch      = "main"
  build_command       = "npm run build"
  published_directory = "dist"
  node_version        = "20"
}
