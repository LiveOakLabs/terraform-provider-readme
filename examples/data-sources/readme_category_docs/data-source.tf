# The "readme_category_docs" data source retrieves a list of docs for a category.
data "readme_category_docs" "example" {
    slug = "example"
}

output "category_docs" {
    value = data.readme_category_docs.example
}