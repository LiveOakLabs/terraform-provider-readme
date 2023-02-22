# The data source retrieves metadata about the current project associated with
# the API token.
data "readme_project" "example" {}

# Output all project metadata.
# This output is redacted because the jwtSecret is marked sensitive.
output "example_project" {
    value     = data.readme_project.example
    sensitive = true
}

# Output a single attribute
output "example_project_name" {
    value = data.readme_project.example.name
}

output "example_project_base_url" {
    value = data.readme_project.example.base_url
}
