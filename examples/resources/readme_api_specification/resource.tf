# Create an API specification resource.
resource "readme_api_specification" "example" {
  # 'definition' accepts a string of an OpenAPI specification definition JSON.
  definition = file("petstore.json")
}

# Output the ID of the created resource.
output "created_spec_id" {
  value = readme_api_specification.example.id
}

# Output the specification JSON of the created resource.
output "created_spec_json" {
  value = readme_api_specification.example.definition
}
