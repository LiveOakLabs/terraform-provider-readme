# Manage a changelog on ReadMe.
resource "readme_changelog" "example" {
  # title can be specified as an attribute or in the body front matter.
  title = "My First Test"
  type = "added"

  # hidden can be specified as an attribute or in the body front matter.
  hidden = false

  # body can be read from a file using Terraform's `file()` or `templatefile()` functions.
  body = "This is my first test. I hope it works!\n\n"
}

# data "readme_changelog" "example" {
#   slug = readme_changelog.example.slug
# }
#
# output "example" {
#   value = data.readme_changelog.example
# }
