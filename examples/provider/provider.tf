terraform {
  required_providers {
    sevalla = {
      source  = "sevalla/sevalla"
      version = "~> 1.2"
    }
  }
}

provider "sevalla" {
  # API key can be set via the SEVALLA_API_KEY environment variable
  # api_key = var.sevalla_api_key
}
