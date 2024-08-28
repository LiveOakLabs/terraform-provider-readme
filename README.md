# Terraform Provider for ReadMe.com

[![Version](https://img.shields.io/github/v/release/liveoaklabs/terraform-provider-readme)](https://github.com/liveoaklabs/terraform-provider-readme/releases)

<img align="right" width="200" src=".github/readme/lob-logo.png">

📖 Refer to <https://registry.terraform.io/providers/LiveOakLabs/readme/latest/docs>
for the latest provider documentation.

☁️ Also see our [Go Client for the ReadMe.com API](https://github.com/liveoaklabs/readme-api-go-client)
that this provider uses.

_This provider is developed by [Live Oak Bank](https://liveoakbank.com) and is
not officially associated with ReadMe.com._

## Getting Started

__Terraform >= 1.0 is required.__

### Configure the Provider

```terraform
provider "readme" {
  # Set the API token here or with the README_API_TOKEN env var.
  api_token = "YOUR_API_TOKEN"
}

terraform {
  required_providers {
    readme = {
      source  = "liveoaklabs/readme"
      version = "0.5.0" # Check for the latest version on the Terraform Registry.
    }
  }
}
```

### Manage Resources

Create a version:

```terraform
resource "readme_version" "example" {
  version   = "1.1.0"
  from      = "1.0.0"
  is_hidden = false
}
```

Create an API specification:

```terraform
resource "readme_api_specification" "example" {
  definition = file("petstore.json")
  semver     = readme_version.example.version_clean
}
```

Create a category:

```terraform
resource "readme_category" "example" {
  title   = "My example category"
  type    = "guide"
  version = readme_version.example.version_clean
}
```

Create a doc:

```terraform
resource "readme_doc" "example" {
  # title can be specified as an attribute or in the body front matter.
  title = "My Example Doc"

  # category_slug can be specified as an attribute or in the body front matter.
  category_slug = readme_category.example.slug

  # hidden can be specified as an attribute or in the body front matter.
  hidden = false

  # order can be specified as an attribute or in the body front matter.
  order = 99

  # type can be specified as an attribute or in the body front matter.
  type = "basic"

  # body can be read from a file using Terraform's `file()` function.
  body = file("mydoc.md")

  version = readme_version.example.version_clean
}
```

### Use Data Sources

The provider includes several data sources. Refer to the
[provider docs on the Terraform registry](https://registry.terraform.io/providers/LiveOakLabs/readme/latest/docs/data-sources/api_registry)
for a full list with examples.

## Disclaimer About Versioning and Development Status

⚠️ This project is currently under active development and is versioned using
the `0.x.x` scheme.

Breaking changes will likely occur and will trigger a minor version increment
(e.g., `0.2.0->0.3.0`).

Users are encouraged to pin the provider to a specific patch version for
maximum stability throughout the `0.x.x` series.

For example:

```hcl
terraform {
  required_providers {
    readme = {
      source  = "liveoaklabs/readme"
      # Pinning to a specific patch version.
      version = "0.5.0"

      # Alternatively, allow for patch updates.
      # version = "~> 0.5.0"
    }
  }
}
```

A stable `1.x` release is planned for the future once the project meets
certain criteria for feature completeness and stability.

Refer to the [CHANGELOG](CHANGELOG.md) for release details.

## Contributing

Refer to [`CONTRIBUTING.md`](CONTRIBUTING.md) for information on contributing to this project.

## License

This project is licensed under the MIT License - see the [`LICENSE`](LICENSE) file for details.
