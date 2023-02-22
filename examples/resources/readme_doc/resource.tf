# Manage docs on ReadMe.
resource "readme_doc" "example" {
    # title can be specified as an attribute or in the body front matter.
    title = "My Example Doc"

    # category can be specified as an attribute or in the body front matter.
    # Use the `readme_category` resource to manage categories.
    category = "633c5a54187d2c008e2e074c"

    # category_slug can be specified as an attribute or in the body front matter.
    # category_slug = "foo-bar"

    # hidden can be specified as an attribute or in the body front matter.
    hidden = false

    # order can be specified as an attribute or in the body front matter.
    order = 99

    # type can be specified as an attribute or in the body front matter.
    type = "basic"

    # body can be read from a file using Terraform's `file()` function.
    # For best results, wrap the string with the `chomp()` function to remove
    # trailing newlines. ReadMe's API trims these implicitly.
    body = chomp(file("mydoc.md"))
}