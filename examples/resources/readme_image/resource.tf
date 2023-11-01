# Upload an image to ReadMe.
resource "readme_image" "example" {
  source = "example.png"
}

output "image_info" {
  value = readme_image.example
}

