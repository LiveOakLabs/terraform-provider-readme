# The "readme_version" data source retrieves metadata about a version on ReadMe.com.
data "readme_version" "example" {
  # Use the 'version' or 'version_clean' attributes to look up a version.
  # For best results, use 'version_clean'.
  # version = "1.0.1"
  version_clean = "1.0.1"
}

# Output all version metadata.
output "example_version" {
  value = data.readme_version.example
}
