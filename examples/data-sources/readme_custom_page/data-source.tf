# Retrieve a single custom page by its slug.
data "readme_custom_page" "example" {
    slug = "my-example-custom-page"
}

output "example" {
    value = data.readme_custom_page.example
}