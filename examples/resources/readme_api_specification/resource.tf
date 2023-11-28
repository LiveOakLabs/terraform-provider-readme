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
# output "created_spec_json" {
#   value = readme_api_specification.example.definition
# }

# ---------------------------------------------------------------------------
# Example of associating a doc resource with the API specification's default
# doc that is automatically created.
resource "readme_doc" "example" {
  # This will be the visible name of the API specification's default doc.
  title = "store"

  # Use the API specification's category ID.
  category = readme_api_specification.example.category.id

  # Specify the slug of the created API spec doc. This is the slug of a
  # specification tag.
  use_slug = "store"

  order = 10
  type  = "basic"
  body  = "This is the Pet Store API specification's default doc."
}
