terraform {
  required_providers {
    sevalla = {
      source  = "sevalla-hosting/sevalla"
      version = "~> 0.1"
    }
  }
}

provider "sevalla" {
  # API key can be set via the SEVALLA_API_KEY environment variable
  # api_key = var.sevalla_api_key
}
