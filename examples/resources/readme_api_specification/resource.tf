# Create an API specification resource.
resource "readme_api_specification" "example" {
  # 'definition' accepts a string of an OpenAPI specification definition JSON.
  definition = file("petstore.json")

  # When an API specification is created, a category is also created but is
  # not deleted when the API specification is deleted. Set this parameter to
  # true to delete the category when the API specification is deleted.
  delete_category = true
}

# Output the ID of the created resource.
output "created_spec_id" {
  value = readme_api_specification.example.id
}

# Output the specification JSON of the created resource.
output "created_spec_json" {
  value = readme_api_specification.example.definition
}
