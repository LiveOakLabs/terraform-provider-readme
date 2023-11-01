# Retrieve a doc from ReadMe.
data "readme_doc" "example" {
  slug = "my-example-doc"
}

output "example_doc" {
  value = data.readme_doc.example
}

