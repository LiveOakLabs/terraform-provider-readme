# The "readme_category" data source retrieves a single category's metadata.
data "readme_category" "example" {
    slug = "example"
}

output "category" {
    value = data.readme_category.example
}