# Manage a custom page on ReadMe.
resource "readme_custom_page" "example" {
  # title can be specified as an attribute or in the body front matter.
  title = "My Example Custom Page"

  # hidden can be specified as an attribute or in the body front matter.
  hidden = false

  # body can be read from a file using Terraform's `file()` or `templatefile()` functions.
  body = file("my-custom-page.md")
}

# Example using HTML.
resource "readme_custom_page" "example_html" {
  title     = "My Example Custom Page"
  html_mode = true
  html      = file("my-custom-page.html")
}
