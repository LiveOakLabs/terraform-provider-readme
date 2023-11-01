# Search for docs on ReadMe.
data "readme_doc_search" "example" {
  query = "*"
}

output "example_doc_search" {
  value = data.readme_doc_search.example
}

