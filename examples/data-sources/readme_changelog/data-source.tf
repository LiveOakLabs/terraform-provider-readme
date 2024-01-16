# Retrieve a single changelog by its slug.
data "readme_changelog" "example" {
  slug = "my-example-changelog"
}

output "example" {
  value = data.readme_changelog.example
}

