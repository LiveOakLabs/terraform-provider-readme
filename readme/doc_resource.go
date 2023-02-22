package readme

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &docResource{}
	_ resource.ResourceWithConfigure   = &docResource{}
	_ resource.ResourceWithImportState = &docResource{}
)

// docResource is the data source implementation.
type docResource struct {
	client *readme.Client
}

// NewDocResource is a helper function to simplify the provider implementation.
func NewDocResource() resource.Resource {
	return &docResource{}
}

// Metadata returns the data source type name.
func (r *docResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_doc"
}

// Configure adds the provider configured client to the data source.
func (r *docResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*readme.Client)
}

// ValidateConfig is used for validating attribute values.
func (r docResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data docModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// category or category_slug must be set. If the attributes aren't set, check the body front matter.
	if data.Category.IsNull() && data.CategorySlug.IsNull() {
		// check front matter for 'category'.
		categoryMatter, diag := valueFromFrontMatter(ctx, data.Body.ValueString(), "Category")
		if diag != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("category"),
				"Error checking front matter during validation.",
				diag,
			)

			return
		}

		// check front matter for 'category_slug'.
		categorySlugMatter, diag := valueFromFrontMatter(
			ctx,
			data.Body.ValueString(),
			"CategorySlug",
		)
		if diag != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("category_slug"),
				"Error checking front matter during validation.",
				diag,
			)

			return
		}

		// Fail if neither category or categorySlug are set in front matter or with attributes.
		if categoryMatter == (reflect.Value{}) && categorySlugMatter == (reflect.Value{}) {
			resp.Diagnostics.AddAttributeError(
				path.Root("category"),
				"Missing required attribute.",
				"category or category_slug must be set. These can be set using the attribute or in the body front matter.",
			)

			return
		}
	}
}

// docPlanToParams maps plan attributes to a `readme.DocParams` struct to create or update a doc.
func docPlanToParams(ctx context.Context, plan docModel) readme.DocParams {
	return readme.DocParams{
		Body:          plan.Body.ValueString(),
		Category:      plan.Category.ValueString(),
		CategorySlug:  plan.CategorySlug.ValueString(),
		Hidden:        boolPoint(plan.Hidden.ValueBool()),
		Order:         intPoint(int(plan.Order.ValueInt64())),
		ParentDoc:     plan.ParentDoc.ValueString(),
		ParentDocSlug: plan.ParentDocSlug.ValueString(),
		Title:         plan.Title.ValueString(),
		Type:          plan.Type.ValueString(),
	}
}

// Create creates the doc and sets the initial Terraform state.
func (r *docResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan.
	var state, plan docModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(plan.Version)
	tflog.Info(ctx, fmt.Sprintf("creating doc with request options=%+v", requestOpts))

	// If a parent doc is set, verify that it exists.
	if plan.VerifyParentDoc.IsNull() || plan.VerifyParentDoc.ValueBool() {
		validParent, detail := r.docValidParent(ctx, plan, requestOpts)
		if !validParent {
			resp.Diagnostics.AddError("Unable to create doc.", detail)

			return
		}
	}

	// Create the doc.
	response, apiResponse, err := r.client.Doc.Create(docPlanToParams(ctx, plan), requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create doc.", clientError(err, apiResponse))

		return
	}

	// Get the doc.
	state, err = getDoc(r.client, ctx, response.Slug, plan, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create doc.",
			fmt.Sprintf(
				"There was a problem retrieving the doc '%s' after creation: %s.",
				response.Slug,
				err.Error(),
			),
		)

		return
	}

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Unable to refresh doc after creation.",
			"There was a problem setting the state.")

		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *docResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state.
	var plan, state docModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(plan.Version)
	tflog.Info(ctx, fmt.Sprintf("retrieving doc with request options=%+v", requestOpts))

	// Get the doc.
	state, err := getDoc(r.client, ctx, state.Slug.ValueString(), plan, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read doc.", err.Error())

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Unable to refresh doc.",
			"There was a problem setting the state.",
		)

		return
	}
}

// Update updates the Doc and sets the updated Terraform state on success.
func (r *docResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan and current state.
	var config, plan, state docModel
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(plan.Version)
	tflog.Info(ctx, fmt.Sprintf("updating doc with request options=%+v", requestOpts))

	// If a parent doc is set, verify that it exists.
	if plan.VerifyParentDoc.IsNull() || plan.VerifyParentDoc.ValueBool() {
		validParent, detail := r.docValidParent(ctx, plan, requestOpts)
		if !validParent {
			resp.Diagnostics.AddError("Unable to update doc.", detail)

			return
		}
	}

	// Update the doc.
	params := docPlanToParams(ctx, plan)
	response, apiResponse, err := r.client.Doc.Update(state.Slug.ValueString(), params, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update doc.", clientError(err, apiResponse))

		return
	}

	// Get the doc.
	plan, err = getDoc(r.client, ctx, response.Slug, plan, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update doc.",
			fmt.Sprintf(
				"There was a problem retrieving the doc '%s' after update: %s.",
				response.Slug,
				err.Error(),
			),
		)

		return
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the Doc and removes the Terraform state on success.
func (r *docResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state.
	var state docModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(state.Version)
	tflog.Info(ctx, fmt.Sprintf("deleting doc with request options=%+v", requestOpts))

	// Delete the doc.
	_, apiResponse, err := r.client.Doc.Delete(state.Slug.ValueString(), requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete doc", clientError(err, apiResponse))
	}
}

// ImportState imports a doc by its slug.
func (r *docResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Import by slug.
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}

// docValidParent verifies that a parent doc exists if the `parent_doc` or `parent_doc_slug` attributes are set.
//
// If neither attribute is set, this skips any evaluation and returns true with an empty string.
//
// If the attributes are set, attempt to retrieve the doc. If an error is returned when retrieving the doc,
// return false and the error as a string for use with the plugin framework's "diag" errors.
//
// If the attributes are set and the doc does not exist, return false and a string stating that it doesn't exist
// for use in the response diag error.
func (r *docResource) docValidParent(
	ctx context.Context,
	plan docModel,
	options readme.RequestOptions,
) (bool, string) {
	if plan.ParentDoc.ValueString() != "" {
		attrVal := "id:" + plan.ParentDoc.ValueString()
		_, _, err := r.client.Doc.Get(attrVal, options)
		if err != nil {
			return false,
				fmt.Sprintf(`Could not find parent_doc matching "%s" (is it hidden?)`+
					"\n"+`For best results, use the "parent_doc_slug" attribute or set `+
					`"verify_parent_doc" to false.`, attrVal,
				)
		}
	} else if plan.ParentDocSlug.ValueString() != "" {
		attrVal := plan.ParentDocSlug.ValueString()
		_, _, err := r.client.Doc.Get(attrVal, options)
		if err != nil {
			return false, fmt.Sprintf("Could not find parent_doc_slug matching %s", attrVal)
		}
	}

	return true, ""
}

// Schema for the readme_doc resource.
func (r *docResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manage docs on ReadMe.com\n\n" +
			"See <https://docs.readme.com/main/reference/getdoc> for more information about this API endpoint.\n\n" +
			"All of the optional attributes except `body` may alternatively be set in the body's front matter. " +
			"Attributes take precedence over values set in front matter.\n\n" +
			"Refer to <https://docs.readme.com/main/docs/rdme> for more information about using front matter in " +
			"ReadMe docs.",
		Attributes: map[string]schema.Attribute{
			"algolia": schema.SingleNestedAttribute{
				Description: "Metadata about the Algolia search integration. " +
					"See <https://docs.readme.com/main/docs/search> for more information.",
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"record_count": schema.Int64Attribute{
						Computed: true,
					},
					"publish_pending": schema.BoolAttribute{
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
						Description: "",
						Computed:    true,
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
				Description: "The body content of the doc, formatted in ReadMe or GitHub flavored Markdown. " +
					"Accepts long page content, for example, greater than 100k characters.",
				Computed: true,
				Optional: true,
			},
			"body_html": schema.StringAttribute{
				Description: "The body content in HTML.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "**Required**. The category ID of the doc. Note that changing the category will result " +
					"in a replacement of the doc resource. This attribute may optionally be set in the body front " +
					"matter or with the `category_slug` attribute.\n\n" +
					"Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.",
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{},
				PlanModifiers: []planmodifier.String{
					fromMatterString("Category"),
				},
			},
			"category_slug": schema.StringAttribute{
				Description: "**Required**. The category ID of the doc. Note that changing the category will result " +
					"in a replacement of the doc resource. This attribute may optionally be set in the body front " +
					"matter with the `categorySlug` key or with the `category` attribute.\n\n" +
					"Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("CategorySlug"),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the version was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deprecated": schema.BoolAttribute{
				Description: "Toggles if a doc is deprecated or not.",
				Computed:    true,
			},
			"error": schema.SingleNestedAttribute{
				Description: "Error code configuration for a doc. This attribute may be set in the body front matter.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Description: "",
						Optional:    true,
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
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					fromMatterBool("Hidden"),
				},
			},
			"icon": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the doc.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
						Description: "",
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
						Computed:    true,
						Description: "List of 'next' page configurations.",
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
				Description: "The position of the doc in the project sidebar. " +
					"This attribute may be set in the body front matter.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					fromMatterInt64("Order"),
				},
			},
			"parent_doc": schema.StringAttribute{
				Description: "For a subpage, specify the parent doc ID." +
					"This attribute may be set in the body front matter with the `parentDoc` key." +
					"The provider cannot verify that a `parent_doc` exists if it is hidden. To " +
					"use a `parent_doc` ID without verifying, set the `verify_parent_doc` " +
					"attribute to `false`.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("ParentDoc"),
				},
			},
			"parent_doc_slug": schema.StringAttribute{
				Description: "For a subpage, specify the parent doc slug instead of the ID." +
					"This attribute may be set in the body front matter with the `parentDocSlug` key." +
					"If a value isn't specified but `parent_doc` is, the provider will attempt to populate this " +
					"value using the `parent_doc` ID unless `verify_parent_doc` is set to `false`.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("ParentDocSlug"),
				},
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
			},
			"slug_updated_at": schema.StringAttribute{
				Description: "The timestamp of when the doc's slug was last updated.",
				Computed:    true,
			},
			"sync_unique": schema.StringAttribute{
				Computed: true,
			},
			"title": schema.StringAttribute{
				Description: "**Required.** The title of the doc." +
					"This attribute may optionally be set in the body front matter.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("Title"),
				},
			},
			"type": schema.StringAttribute{
				Description: `**Required.** Type of the doc. The available types all show up under the /docs/ URL ` +
					`path of your docs project (also known as the "guides" section). Can be "basic" (most common), ` +
					`"error" (page desribing an API error), or "link" (page that redirects to an external link).` +
					"This attribute may optionally be set in the body front matter.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					fromMatterString("Type"),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "The timestamp of when the doc was last updated.",
				Computed:    true,
			},
			"user": schema.StringAttribute{
				Description: "The ID of the author of the doc in the web editor.",
				Computed:    true,
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
			"verify_parent_doc": schema.BoolAttribute{
				Description: "Enables or disables the provider verifying the `parent_doc` exists. When using the " +
					"`parent_doc` attribute with a hidden parent, the provider is unable to verify if the parent " +
					"exists. Setting this to `false` will disable this behavior. When `false`, the `parent_doc_slug` " +
					"value will not be resolved by the provider unless explicitly set. The `parent_doc_slug` " +
					"attribute may be used as an alternative. Verifying a `parent_doc` by ID does not work if the " +
					"parent is hidden.",
				Optional: true,
			},
		},
	}
}
