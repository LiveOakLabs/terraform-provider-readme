# Manage a changelog on ReadMe.
resource "readme_changelog" "example" {
  # title and type can be specified as an attribute or in the body front matter.
  title  = "My Example Changelog"
  type   = "added"
  hidden = false

  # body can be read from a file using Terraform's `file()` or `templatefile()` functions.
  body = "* Added support for foo\n* Added support for bar"
}
