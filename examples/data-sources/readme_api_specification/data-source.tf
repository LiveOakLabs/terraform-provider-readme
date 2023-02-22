# Retrieve a specific API specification.
data "readme_api_specification" "example" {
  # lookup by id. id is unique
  id = readme_api_specification.example.id
  # lookup by title. title is not unique. the last matched is returned.
  # title = "Awesome New API"
}

# Output the title from the data source.
output "api_spec_data_lookup" {
  value = data.readme_api_specification.example.title
}
