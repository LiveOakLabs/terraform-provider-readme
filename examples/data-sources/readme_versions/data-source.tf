# The "readme_versions" data source retrieves a list of versions and returns summarized metadata.
data "readme_versions" "example" {}

# Return the full list.
output "example_versions" {
  value = data.readme_versions.example.versions
}

# Retrieve a specific attribute from version in the list.
output "example_version_detail" {
  value = tolist(data.readme_versions.example.versions)[0].version
}
