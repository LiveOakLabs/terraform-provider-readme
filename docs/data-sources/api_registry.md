---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "readme_api_registry Data Source - readme"
subcategory: ""
description: |-
  Retrieve an API specification definition from the API registry on ReadMe.com
  See https://docs.readme.com/main/reference/getapiregistry for more information about this API endpoint.
---

# readme_api_registry (Data Source)

Retrieve an API specification definition from the API registry on ReadMe.com

See <https://docs.readme.com/main/reference/getapiregistry> for more information about this API endpoint.

## Example Usage

```terraform
# The data source retrieves an API specification from the registry.
#
# Reference: https://docs.readme.com/main/reference/getapiregistry
data "readme_api_registry" "example" {
  # An API Registry UUID. This can be found by navigating to your API Reference
  # page and viewing code snippets for Node with the API library.
  uuid = ""
}

# Output the API specification from the registry.
output "example_api_registry" {
  value = data.readme_api_registry.example.definition
}

output "example_registry_uuid" {
  value = data.readme_api_registry.example.uuid
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `uuid` (String) The UUID of an API registry definition.

### Read-Only

- `definition` (String) The raw JSON definition of an API specification.
- `id` (String) The internal ID of this resource.
