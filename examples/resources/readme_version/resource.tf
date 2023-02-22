# The "readme_version" resource manages versions in ReadMe.
# Create a version resource.
resource "readme_version" "example" {
  version   = "1.1.0"
  from      = "1.0.0"
  is_hidden = true
}

# Output the created resource attributes.
output "created_version" {
  value = readme_version.example
}

