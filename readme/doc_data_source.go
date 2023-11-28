package readme

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &docDataSource{}
	_ datasource.DataSourceWithConfigure = &docDataSource{}
)

// docDataSource is the data source implementation.
type docDataSource struct {
	client *readme.Client
}

// NewDocDataSource is a helper function to simplify the provider implementation.
func NewDocDataSource() datasource.DataSource {
	return &docDataSource{}
}

// Metadata returns the data source type name.
func (d *docDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_doc"
}

// Read refreshes the Terraform state with the latest data.
func (d *docDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state docModel

	// Get config.
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(state.Version)
	tflog.Info(ctx, fmt.Sprintf("retrieving doc with request options=%+v", requestOpts))

	// Get the doc.
	state, _, err := getDoc(d.client, ctx, state.Slug.ValueString(), state, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve doc metadata.", err.Error())

		return
	}

	state.Body = state.BodyClean

	// Set state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *docDataSource) Configure(
	ctx context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*readme.Client)
}

// Schema for the readme_doc data source.
func (d *docDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	// nolint: goconst // Attribute descriptions are repeated across doc
	// resources and data sources.
	resp.Schema = schema.Schema{
		Description: "Retrieve docs on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getdoc> for more information about this API endpoint.\n\n",
		Attributes: map[string]schema.Attribute{
			"algolia": schema.SingleNestedAttribute{
				Description: "Metadata about the Algolia search integration. " +
					"See <https://docs.readme.com/main/docs/search> for more information.",
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"publish_pending": schema.BoolAttribute{
						Computed: true,
					},
					"record_count": schema.Int64Attribute{
						Computed: true,
					},
					"updated_at": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"api": schema.SingleNestedAttribute{
				Description: "Metadata for an API doc.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"api_setting": schema.StringAttribute{
						Computed: true,
					},
					"auth": schema.StringAttribute{
						Computed: true,
					},
					"examples": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"codes": schema.ListNestedAttribute{
								Computed: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"code": schema.StringAttribute{
											Computed: true,
										},
										"language": schema.StringAttribute{
											Computed: true,
										},
									},
								},
							},
						},
					},
					"method": schema.StringAttribute{
						Computed: true,
					},
					"params": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"default": schema.StringAttribute{
									Computed: true,
								},
								"desc": schema.StringAttribute{
									Computed: true,
								},
								"enum_values": schema.StringAttribute{
									Computed: true,
								},
								"id": schema.StringAttribute{
									Computed: true,
								},
								"in": schema.StringAttribute{
									Computed: true,
								},
								"name": schema.StringAttribute{
									Computed: true,
								},
								"ref": schema.StringAttribute{
									Computed: true,
								},
								"required": schema.BoolAttribute{
									Computed: true,
								},
								"type": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
					"results": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"codes": schema.ListNestedAttribute{
								Computed: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"code": schema.StringAttribute{
											Computed: true,
										},
										"language": schema.StringAttribute{
											Computed: true,
										},
										"name": schema.StringAttribute{
											Computed: true,
										},
										"status": schema.Int64Attribute{
											Computed: true,
										},
									},
								},
							},
						},
					},
					"url": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"body": schema.StringAttribute{
				Description: "The body content of the doc, formatted in ReadMe or GitHub flavored Markdown.",
				Computed:    true,
			},
			"body_clean": schema.StringAttribute{
				Description: "The body content of the doc, formatted in ReadMe or GitHub flavored Markdown. " +
					"This is an alias for the `body` attribute.",
				Computed: true,
			},
			"body_html": schema.StringAttribute{
				Description: "The body content in HTML.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category ID of the doc. Note that changing the category will result in a " +
					"replacement of the doc resource.",
				Computed: true,
			},
			"category_slug": schema.StringAttribute{
				Description: "**Required**. The category ID of the doc. Note that changing the category will result " +
					"in a replacement of the doc resource. This attribute may optionally be set in the body front matter.",
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the version was created.",
				Computed:    true,
			},
			"deprecated": schema.BoolAttribute{
				Description: "Toggles if a doc is deprecated or not.",
				Computed:    true,
			},
			"error": schema.SingleNestedAttribute{
				Description: "Error code configuration for a doc. This attribute may be set in the body front matter.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"excerpt": schema.StringAttribute{
				Description: "A short summary of the content.",
				Computed:    true,
			},
			"hidden": schema.BoolAttribute{
				Description: "Toggles if a doc is hidden or not. This attribute may be set in the body front matter.",
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the doc.",
				Computed:    true,
			},
			"is_api": schema.BoolAttribute{
				Computed: true,
			},
			"is_reference": schema.BoolAttribute{
				Computed: true,
			},
			"link_external": schema.BoolAttribute{
				Computed: true,
			},
			"link_url": schema.StringAttribute{
				Computed: true,
			},
			"metadata": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Computed: true,
					},
					"image": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
					},
					"title": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"next": schema.SingleNestedAttribute{
				Description: "Information about the 'next' pages in a series.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Computed: true,
					},
					"pages": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed: true,
								},
								"type": schema.StringAttribute{
									Computed: true,
								},
								"icon": schema.StringAttribute{
									Computed: true,
								},
								"slug": schema.StringAttribute{
									Computed: true,
								},
								"category": schema.StringAttribute{
									Computed: true,
								},
								"deprecated": schema.BoolAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
			"order": schema.Int64Attribute{
				Description: "The position of the doc in the project sidebar.",
				Computed:    true,
			},
			"parent_doc": schema.StringAttribute{
				Description: "If the doc has a parent doc, this is doc ID of the parent.",
				Computed:    true,
			},
			"parent_doc_slug": schema.StringAttribute{
				Description: "If the doc has a parent doc, this is doc slug of the parent.",
				Computed:    true,
			},
			"previous_slug": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Description: "The ID of the project the doc is in.",
				Computed:    true,
			},
			"revision": schema.Int64Attribute{
				Description: "A number that is incremented upon doc updates.",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the doc.",
				Computed:    true,
				Optional:    true,
			},
			"slug_updated_at": schema.StringAttribute{
				Description: "The timestamp of when the doc's slug was last updated.",
				Computed:    true,
			},
			"sync_unique": schema.StringAttribute{
				Computed: true,
			},
			"title": schema.StringAttribute{
				Description: "The title of the doc.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: `Type of the doc. The available types all show up under the /docs/ URL path of your ` +
					`docs project (also known as the "guides" section). Can be "basic" (most common), "error" (page ` +
					`desribing an API error), or "link" (page that redirects to an external link).`,
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp of when the doc was last updated.",
				Computed:    true,
			},
			"user": schema.StringAttribute{
				Description: "The ID of the author of the doc in the web editor.",
				Computed:    true,
			},
			// This isn't used by the doc data source, but must be present because the struct
			// is shared with the doc resource, which does use it.
			// In the future, we may want to split the struct into separate types for the
			// resource and data source.
			"use_slug": schema.StringAttribute{
				Description: "This is an unused attribute in the data source that is present to " +
					"satisfy the model shared with the doc resource. It may be removed in the future.",
				Computed: true,
			},
			"verify_parent_doc": schema.BoolAttribute{
				Description: "Enables or disables the provider verifying the `parent_doc` exists. When using the " +
					"`parent_doc` attribute with a hidden parent, the provider is unable to verify if the parent " +
					"exists. Setting this to `false` will disable this behavior. When `false`, the `parent_doc_slug` " +
					"value will not be resolved by the provider unless explicitly set. The `parent_doc_slug` " +
					"attribute may be used as an alternative. Verifying a `parent_doc` by ID does not work if the " +
					"parent is hidden.",
				Optional: true,
			},
			"version": schema.StringAttribute{
				Description: "The version to create the doc under.",
				Optional:    true,
				Computed:    true,
			},
			"version_id": schema.StringAttribute{
				Description: "The version ID the doc is associated with.",
				Computed:    true,
			},
		},
	}
}
