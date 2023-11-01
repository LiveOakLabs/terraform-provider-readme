provider "readme" {
  # Set the API URL here or with the README_API_URL env var. It is optional and
  # will use the library's default URL if unset.
  # api_url   = ""

  # Set the API token here or with the README_API_TOKEN env var.
  # api_token = ""
}

terraform {
  required_providers {
    readme = {
      source = "liveoaklabs/readme"
    }
  }
}

