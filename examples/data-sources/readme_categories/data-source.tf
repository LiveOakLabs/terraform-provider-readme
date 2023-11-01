# The "readme_categories" data source retrieves a list of all categories for a ReadMe project.
data "readme_categories" "example" {}

output "categories" {
  value = data.readme_categories.example
}

