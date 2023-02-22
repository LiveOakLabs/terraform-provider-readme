# The "readme_category" resource manages the lifecycle of a category in ReadMe.
resource "readme_category" "example" {
    title = "My example category"
    type  = "guide"
}