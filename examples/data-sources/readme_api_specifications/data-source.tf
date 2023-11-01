# Return all API specifications in a ReadMe project.
data "readme_api_specifications" "all" {}

# Return specifications with a matching category title.
data "readme_api_specifications" "filter" {
  filter = {
    category_title = [
      "Test API specification 1",
      "Test API specification 2",
    ]
  }
}

# Return specifications that have a category (specs that are visible).
data "readme_api_specifications" "filter2" {
  filter = {
    has_category = true
  }
}

# Return specfications that match any filter.
# Any specification that matches any of the category_id or category_title AND
# has a category will be returned.
data "readme_api_specifications" "filter3" {
  filter = {
    category_id = [
      "5f7f9c1b0f1c9a001e1b1b1a",
      "60a1b1b0f1c9a001e1b1b1b"
    ]

    category_title = [
      "Test API specification 1",
      "Test API specification 2",
    ]

    category_slug = [
      "test-api-specification-3",
      "test-api-specification-4",
    ]

    has_category = true
  }
}

# Return specifications that match a title filter.
data "readme_api_specifications" "filter4" {
  filter = {
    title = [
      "Test API specification 1",
      "Test API specification 2",
    ]
  }
}

