# The data source retrieves an API specification from the registry.
#
# Reference: https://docs.readme.com/main/reference/getapiregistry
data "readme_api_registry" "example" {
  # An API Registry UUID. This can be found by navigating to your API Reference
  # page and viewing code snippets for Node with the API library.
  uuid = ""
}

# Output the API specification from the registry.
output "example_api_registry" {
    value = data.readme_api_registry.example.definition
}

output "example_registry_uuid" {
    value = data.readme_api_registry.example.uuid
}
