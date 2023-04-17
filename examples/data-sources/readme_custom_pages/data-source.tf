# Retrieve all custom pages.
data "readme_custom_pages" "example" {}

output "example" {
    value = data.readme_custom_pages.example
}