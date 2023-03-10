---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "readme_doc Resource - readme"
subcategory: ""
description: |-
  Manage docs on ReadMe.com
  See https://docs.readme.com/main/reference/getdoc for more information about this API endpoint.
  All of the optional attributes except body may alternatively be set in the body's front matter. Attributes take precedence over values set in front matter.
  Refer to https://docs.readme.com/main/docs/rdme for more information about using front matter in ReadMe docs.
---

# readme_doc (Resource)

Manage docs on ReadMe.com

See <https://docs.readme.com/main/reference/getdoc> for more information about this API endpoint.

All of the optional attributes except `body` may alternatively be set in the body's front matter. Attributes take precedence over values set in front matter.

Refer to <https://docs.readme.com/main/docs/rdme> for more information about using front matter in ReadMe docs.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `body` (String) The body content of the doc, formatted in ReadMe or GitHub flavored Markdown. Accepts long page content, for example, greater than 100k characters.
- `category` (String) **Required**. The category ID of the doc. Note that changing the category will result in a replacement of the doc resource. This attribute may optionally be set in the body front matter or with the `category_slug` attribute.

Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.
- `category_slug` (String) **Required**. The category ID of the doc. Note that changing the category will result in a replacement of the doc resource. This attribute may optionally be set in the body front matter with the `categorySlug` key or with the `category` attribute.

Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.
- `error` (Attributes) Error code configuration for a doc. This attribute may be set in the body front matter. (see [below for nested schema](#nestedatt--error))
- `hidden` (Boolean) Toggles if a doc is hidden or not. This attribute may be set in the body front matter.
- `order` (Number) The position of the doc in the project sidebar. This attribute may be set in the body front matter.
- `parent_doc` (String) For a subpage, specify the parent doc ID.This attribute may be set in the body front matter with the `parentDoc` key.The provider cannot verify that a `parent_doc` exists if it is hidden. To use a `parent_doc` ID without verifying, set the `verify_parent_doc` attribute to `false`.
- `parent_doc_slug` (String) For a subpage, specify the parent doc slug instead of the ID.This attribute may be set in the body front matter with the `parentDocSlug` key.If a value isn't specified but `parent_doc` is, the provider will attempt to populate this value using the `parent_doc` ID unless `verify_parent_doc` is set to `false`.
- `title` (String) **Required.** The title of the doc.This attribute may optionally be set in the body front matter.
- `type` (String) **Required.** Type of the doc. The available types all show up under the /docs/ URL path of your docs project (also known as the "guides" section). Can be "basic" (most common), "error" (page desribing an API error), or "link" (page that redirects to an external link).This attribute may optionally be set in the body front matter.
- `verify_parent_doc` (Boolean) Enables or disables the provider verifying the `parent_doc` exists. When using the `parent_doc` attribute with a hidden parent, the provider is unable to verify if the parent exists. Setting this to `false` will disable this behavior. When `false`, the `parent_doc_slug` value will not be resolved by the provider unless explicitly set. The `parent_doc_slug` attribute may be used as an alternative. Verifying a `parent_doc` by ID does not work if the parent is hidden.
- `version` (String) The version to create the doc under.

### Read-Only

- `algolia` (Attributes) Metadata about the Algolia search integration. See <https://docs.readme.com/main/docs/search> for more information. (see [below for nested schema](#nestedatt--algolia))
- `api` (Attributes) Metadata for an API doc. (see [below for nested schema](#nestedatt--api))
- `body_html` (String) The body content in HTML.
- `created_at` (String) Timestamp of when the version was created.
- `deprecated` (Boolean) Toggles if a doc is deprecated or not.
- `excerpt` (String) A short summary of the content.
- `icon` (String)
- `id` (String) The ID of the doc.
- `is_api` (Boolean)
- `is_reference` (Boolean)
- `link_external` (Boolean)
- `link_url` (String)
- `metadata` (Attributes) (see [below for nested schema](#nestedatt--metadata))
- `next` (Attributes) Information about the 'next' pages in a series. (see [below for nested schema](#nestedatt--next))
- `previous_slug` (String)
- `project` (String) The ID of the project the doc is in.
- `revision` (Number) A number that is incremented upon doc updates.
- `slug` (String) The slug of the doc.
- `slug_updated_at` (String) The timestamp of when the doc's slug was last updated.
- `sync_unique` (String)
- `updated_at` (String) The timestamp of when the doc was last updated.
- `user` (String) The ID of the author of the doc in the web editor.
- `version_id` (String) The version ID the doc is associated with.

<a id="nestedatt--error"></a>
### Nested Schema for `error`

Optional:

- `code` (String)


<a id="nestedatt--algolia"></a>
### Nested Schema for `algolia`

Read-Only:

- `publish_pending` (Boolean)
- `record_count` (Number)
- `updated_at` (String)


<a id="nestedatt--api"></a>
### Nested Schema for `api`

Read-Only:

- `api_setting` (String)
- `auth` (String)
- `examples` (Attributes) (see [below for nested schema](#nestedatt--api--examples))
- `method` (String)
- `params` (Attributes List) (see [below for nested schema](#nestedatt--api--params))
- `results` (Attributes) (see [below for nested schema](#nestedatt--api--results))
- `url` (String)

<a id="nestedatt--api--examples"></a>
### Nested Schema for `api.examples`

Read-Only:

- `codes` (Attributes List) (see [below for nested schema](#nestedatt--api--examples--codes))

<a id="nestedatt--api--examples--codes"></a>
### Nested Schema for `api.examples.codes`

Read-Only:

- `code` (String)
- `language` (String)



<a id="nestedatt--api--params"></a>
### Nested Schema for `api.params`

Read-Only:

- `default` (String)
- `desc` (String)
- `enum_values` (String)
- `id` (String)
- `in` (String)
- `name` (String)
- `ref` (String)
- `required` (Boolean)
- `type` (String)


<a id="nestedatt--api--results"></a>
### Nested Schema for `api.results`

Read-Only:

- `codes` (Attributes List) (see [below for nested schema](#nestedatt--api--results--codes))

<a id="nestedatt--api--results--codes"></a>
### Nested Schema for `api.results.codes`

Read-Only:

- `code` (String)
- `language` (String)
- `name` (String)
- `status` (Number)




<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

Read-Only:

- `description` (String)
- `image` (List of String)
- `title` (String)


<a id="nestedatt--next"></a>
### Nested Schema for `next`

Read-Only:

- `description` (String)
- `pages` (Attributes List) List of 'next' page configurations. (see [below for nested schema](#nestedatt--next--pages))

<a id="nestedatt--next--pages"></a>
### Nested Schema for `next.pages`

Read-Only:

- `category` (String)
- `deprecated` (Boolean)
- `icon` (String)
- `name` (String)
- `slug` (String)
- `type` (String)


