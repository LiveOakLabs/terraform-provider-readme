package readme

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/liveoaklabs/readme-api-go-client/readme"
	"github.com/liveoaklabs/terraform-provider-readme/readme/frontmatter"
)

const docResourceDesc = `
Manage docs on ReadMe.com

See <https://docs.readme.com/main/reference/getdoc> for more information about this API endpoint.

## Front Matter

Docs on ReadMe support setting some attributes using front matter.
Resource attributes take precedence over front matter attributes in the provider.

Refer to <https://docs.readme.com/main/docs/rdme> for more information about using front matter
in ReadMe docs and custom pages.

## Doc Slugs

Docs in ReadMe are uniquely identified by their slugs. The slug is a URL-friendly string that
is generated upon doc creation. By default, this is a normalized version of the doc title.
The slug cannot be altered using the API or the Terraform Provider, but can be edited in the
ReadMe web UI.

This creates challenges when managing docs with Terraform. To address this, the provider supports
the ` + "`use_slug`" + ` attribute. When set, the provider will attempt to manage an existing
doc by its slug. This can also be set in front matter using the ` + "`slug`" + ` key.

If this attribute is set and the doc does not exist, an error will be returned. This is intended
to be set when inheriting management of an existing doc or when customizing the slug *after*
the doc has been created.

Note that doc slugs are shared between Guides and API Specification References.

⚠️ **Experimental:** The ` + "`use_slug`" + ` attribute is experimental and may result in unexpected
behavior.

## Destroying Docs with Children

Docs in ReadMe can have child docs.
Terraform can infer a doc's relationship when they are all managed by the provider and delete them
in the proper order as normal when referenced appropriately or when using ` + "`depends_on`." + `

However, when managing docs with children, the provider may not be able to infer the relationship
between parent and child docs, particularly in edge cases such as using the ` + "`use_slug`" + `
attribute to manage an API reference's parent doc.

When destroying a doc, the provider will check for child docs and prevent deletion if they exist.
This behavior can be controlled with the ` + "`config.destroy_child_docs`" + ` attribute. When set to
true, the provider will destroy child docs prior to deleting the parent doc. Setting this as a provider
configuration attribute allows for it to be toggled without requiring changes to the resource.

When ` + "`config.destroy_child_docs`" + ` is set to ` + "`true`" + `, the provider will log a
warning when child docs are deleted before the parent doc.

For best results, manage docs with Terraform and set their relationship by referencing the resource
address of the parent doc in the child doc's ` + "`parent_doc_slug`" + ` or ` + "`depends_on`" + `
attributes. This ensures they are deleted in the correct order.
`

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &docResource{}
	_ resource.ResourceWithConfigure   = &docResource{}
	_ resource.ResourceWithImportState = &docResource{}
)

// docResource is the data source implementation.
type docResource struct {
	client *readme.Client
	config providerConfig
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

// Configure adds the provider configured client to the resource.
func (r *docResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	cfg := req.ProviderData.(*providerData)
	r.client = cfg.client
	r.config = cfg.config
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
		categoryMatter, diag := frontmatter.GetValue(ctx, data.Body.ValueString(), "Category")
		if diag != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("category"),
				"Error checking front matter during validation.",
				diag,
			)

			return
		}

		// check front matter for 'category_slug'.
		categorySlugMatter, diag := frontmatter.GetValue(
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

func (r *docResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	plan := &docModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	state := &docModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() || plan == nil {
		return
	}

	if state == nil {
		tflog.Info(ctx, fmt.Sprintf("state is nil for doc %s", plan.Slug.ValueString()))
		plan.BodyClean = types.StringUnknown()
		plan.BodyHTML = types.StringUnknown()
		plan.Revision = types.Int64Unknown()
		plan.UpdatedAt = types.StringUnknown()
		plan.User = types.StringUnknown()
		diags := resp.Plan.Set(ctx, plan)
		resp.Diagnostics.Append(diags...)

		return
	}

	diags := resp.Plan.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	// The 'algolia', 'revision', 'updated_at', and 'user' attributes are
	// volatile and may show changes in the post-apply diff. If other
	// attributes are changed, set these attributes to unknown to trigger a
	// refresh.
	if !state.BodyClean.Equal(plan.BodyClean) ||
		!state.BodyHTML.Equal(plan.BodyHTML) ||
		!state.Category.Equal(plan.Category) ||
		!state.CategorySlug.Equal(plan.CategorySlug) ||
		!state.Hidden.Equal(plan.Hidden) ||
		!state.Order.Equal(plan.Order) ||
		!state.ParentDoc.Equal(plan.ParentDoc) ||
		!state.ParentDocSlug.Equal(plan.ParentDocSlug) ||
		!state.Slug.Equal(plan.Slug) ||
		!state.Title.Equal(plan.Title) ||
		!state.Type.Equal(plan.Type) {

		tflog.Info(ctx, fmt.Sprintf("setting volatile attributes to unknown for doc %s", plan.Slug.ValueString()))

		plan.Revision = types.Int64Unknown()
		plan.UpdatedAt = types.StringUnknown()
		plan.User = types.StringUnknown()

		plan.Algolia = types.ObjectUnknown(
			map[string]attr.Type{
				"record_count":    types.Int64Type,
				"publish_pending": types.BoolType,
				"updated_at":      types.StringType,
			},
		)

	}

	diags = resp.Plan.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// docPlanToParams maps plan attributes to a `readme.DocParams` struct to create or update a doc.
func docPlanToParams(ctx context.Context, plan docModel) readme.DocParams {
	params := readme.DocParams{
		Body:   plan.Body.ValueString(),
		Hidden: plan.Hidden.ValueBoolPointer(),
		Order:  intPoint(int(plan.Order.ValueInt64())),
		Title:  plan.Title.ValueString(),
		Type:   plan.Type.ValueString(),
	}

	// Only use one of Category or CategorySlug.
	if plan.Category.ValueString() != "" {
		params.Category = plan.Category.ValueString()
	} else {
		params.CategorySlug = plan.CategorySlug.ValueString()
	}

	// Only use one of ParentDoc or ParentDocSlug.
	if plan.ParentDoc.ValueString() != "" {
		params.ParentDoc = plan.ParentDoc.ValueString()
	} else {
		params.ParentDocSlug = plan.ParentDocSlug.ValueString()
	}

	return params
}

// Create creates the doc and sets the initial Terraform state.
func (r *docResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var err error
	var doc readme.Doc
	var apiResponse *readme.APIResponse

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

	useSlug := plan.UseSlug.ValueString() != "" && plan.UseSlug.ValueString() != "null"
	exists := false
	if useSlug {
		exists, err = r.docExists(ctx, plan.UseSlug.ValueString(), requestOpts)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create doc.", clientError(err, apiResponse))

			return
		}
	}

	if exists {
		// Adopt the doc.
		adopted, err := r.adoptDoc(ctx, plan, requestOpts)
		if err != nil {
			hint := fmt.Sprintf("\nHint: A value for the `use_slug` attribute is set to '%s', "+
				"but the doc could not be found. Ensure the doc exists and the slug is correct. "+
				"Otherwise, remove the `use_slug` attribute.", plan.UseSlug.ValueString())
			resp.Diagnostics.AddError("Unable to create doc.", "Error: "+err.Error()+hint)

			return
		}
		if adopted == nil {
			resp.Diagnostics.AddError("Unable to create doc.", "adopted doc is nil after successful adoption.")

			return
		}
		doc = *adopted
	} else {
		// Create the doc.
		doc, apiResponse, err = r.client.Doc.Create(docPlanToParams(ctx, plan), requestOpts)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create doc.", clientError(err, apiResponse))

			return
		}
	}

	// Get the doc.
	state, _, err = getDoc(r.client, ctx, doc.Slug, plan, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create doc.",
			fmt.Sprintf("There was a problem retrieving the doc '%s' after creation: %s.", doc.Slug, err.Error()),
		)

		return
	}

	// Get the doc a second time to ensure the state is fully populated.
	state, _, err = getDoc(r.client, ctx, doc.Slug, state, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create doc.",
			fmt.Sprintf("There was a problem retrieving the doc '%s' after creation: %s.", doc.Slug, err.Error()),
		)

		return
	}

	// Set state to fully populated data.
	if state.UseSlug.ValueString() == "" || state.UseSlug.ValueString() == "null" {
		state.UseSlug = types.StringValue(state.Slug.ValueString())
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Unable to refresh doc after creation.",
			"There was a problem setting the state.")

		return
	}
}

func (r *docResource) docExists(
	ctx context.Context,
	slug string,
	options readme.RequestOptions,
) (bool, error) {
	_, _, err := r.client.Doc.Get(slug, options)
	if err != nil {
		return false, err
	}

	return true, nil
}

// adoptDoc attempts to retrieve a doc by its slug and update it with the plan attributes.
// This is used when the `use_slug` attribute is set to assume management of an existing doc.
func (r *docResource) adoptDoc(
	ctx context.Context,
	plan docModel,
	requestOpts readme.RequestOptions,
) (*readme.Doc, error) {
	slug := plan.UseSlug.ValueString()
	tflog.Info(ctx, fmt.Sprintf("using slug %s", slug))
	existing, _, err := getDoc(r.client, ctx, slug, plan, requestOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve doc '%s': %w", slug, err)
	}

	if existing.Slug.ValueString() == "" {
		return nil, fmt.Errorf("doc '%s' not found", slug)
	}

	// Update the existing doc.
	tflog.Info(ctx, fmt.Sprintf("updating doc %s", slug))
	doc, _, err := r.client.Doc.Update(slug, docPlanToParams(ctx, plan), requestOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update doc '%s': %w", slug, err)
	}

	return &doc, nil
}

// Read refreshes the Terraform state with the latest data.
func (r *docResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state.
	var state docModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.Slug.ValueString()
	stateID := state.ID.ValueString()

	if state.UseSlug.ValueString() != "" {
		tflog.Info(ctx, fmt.Sprintf("use_slug is set to %s.", state.UseSlug.ValueString()))
		slug = state.UseSlug.ValueString()
	}

	requestOpts := apiRequestOptions(state.Version)
	logMsg := fmt.Sprintf("retrieving doc %s with request options=%+v", slug, requestOpts)
	tflog.Info(ctx, logMsg)

	// Get the doc.
	state, apiResponse, err := getDoc(r.client, ctx, slug, state, requestOpts)
	if err != nil { // nolint:nestif // TODO: refactor
		if apiResponse != nil && apiResponse.HTTPResponse.StatusCode == 404 {
			// Attempt to find the doc by ID by searching all docs.
			// While the slug is the primary identifier to request a doc, the
			// slug is not stable and can be changed through the web UI.
			tflog.Info(ctx, fmt.Sprintf("doc %s not found when looking up by slug, performing search", slug))
			state, apiResponse, err = getDoc(r.client, ctx, IDPrefix+stateID, state, requestOpts)
			if err != nil {
				if strings.Contains(err.Error(), "no doc found matching id") ||
					strings.Contains(
						err.Error(),
						fmt.Sprintf("The doc with the slug '%s' couldn't be found.", IDPrefix+stateID),
					) {
					tflog.Info(
						ctx,
						fmt.Sprintf(
							"doc %s not found when searching by slug or ID %s, removing from state",
							slug, stateID))
					resp.State.RemoveResource(ctx)

					return
				}
				hint := "Hint: If you changed the doc slug using the web UI, set the `use_slug` " +
					"attribute or the `slug` frontmatter key to the new slug.\n"
				resp.Diagnostics.AddWarning("Unable to search for doc.", hint+clientError(err, apiResponse))

				return
			}
		} else {
			resp.Diagnostics.AddError("Unable to retrieve doc.", clientError(err, apiResponse))

			return
		}
	}

	// Set refreshed state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	var plan, state docModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestOpts := apiRequestOptions(plan.Version)

	// If a parent doc is set, verify that it exists.
	if plan.VerifyParentDoc.IsNull() || plan.VerifyParentDoc.ValueBool() {
		validParent, detail := r.docValidParent(ctx, plan, requestOpts)
		if !validParent {
			resp.Diagnostics.AddError("Unable to update doc.", detail)

			return
		}
	}
	slug := state.Slug.ValueString()

	if state.UseSlug.ValueString() != "" {
		tflog.Info(ctx, fmt.Sprintf("use_slug is set to %s.", state.UseSlug.ValueString()))
		slug = state.UseSlug.ValueString()
	}

	tflog.Info(ctx, fmt.Sprintf("updating doc %s with request options=%+v", slug, requestOpts))

	// Update the doc.
	params := docPlanToParams(ctx, plan)
	response, apiResponse, err := r.client.Doc.Update(slug, params, requestOpts)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update doc.", clientError(err, apiResponse))

		return
	}

	// Get the doc.
	plan, _, err = getDoc(r.client, ctx, response.Slug, plan, requestOpts)
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

	// Get the doc a second time to ensure the state is fully populated.
	plan, _, err = getDoc(r.client, ctx, response.Slug, plan, requestOpts)
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

	// Set state to fully populated data.
	if plan.UseSlug.ValueString() == "" || plan.UseSlug.ValueString() == "null" {
		plan.UseSlug = types.StringValue(plan.Slug.ValueString())
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

	// Check the category's docs to find the doc and its children.
	docs, _, err := r.client.Category.GetDocs(state.CategorySlug.ValueString(), requestOpts)
	if err != nil {
		resp.Diagnostics.AddWarning("Unable to retrieve category docs.", clientError(err, nil))
	}

	// Gather the slugs of the docs to delete, including the parent and its children.
	slugsToDelete, parentDoc := r.identifyDocsToDelete(ctx, state.Slug.ValueString(), docs, resp)
	if parentDoc == nil {
		return
	}

	// Ensure children are handled correctly.
	// Delete them first if DestroyChildDocs is enabled. Otherwise, return.
	if !r.handleChildDocs(ctx, resp, state.Slug.ValueString(), parentDoc) {
		return
	}

	// Perform deletions
	r.deleteDocs(ctx, resp, slugsToDelete, requestOpts)

	// Warn if child docs were deleted
	if len(slugsToDelete) > 1 && r.config.DestroyChildDocs.ValueBool() {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("Doc '%s' and its child docs were destroyed!", state.Slug.ValueString()),
			fmt.Sprintf("The provider configuration 'config.destroy_child_docs' is set to true. "+
				"Child docs were deleted before the parent doc '%s': %s", state.Slug.ValueString(),
				strings.Join(slugsToDelete[1:], ", ")),
		)
	}
}

// identifyDocsToDelete finds the doc and its children to delete.
func (r *docResource) identifyDocsToDelete(
	ctx context.Context,
	slug string,
	docs []readme.CategoryDocs,
	resp *resource.DeleteResponse,
) ([]string, *readme.CategoryDocs) {
	var slugsToDelete []string
	var parentDoc *readme.CategoryDocs

	for _, doc := range docs {
		if parentDoc, slugsToDelete = r.matchDocAndChildren(doc, slug); parentDoc != nil {
			break
		}
	}

	if parentDoc == nil {
		resp.Diagnostics.AddWarning(
			"Doc not found",
			fmt.Sprintf("The doc with slug '%s' was not found in the retrieved category docs.", slug),
		)
	}

	return slugsToDelete, parentDoc
}

// matchDocAndChildren finds the doc and its children in the category docs.
func (r *docResource) matchDocAndChildren(doc readme.CategoryDocs, slug string) (*readme.CategoryDocs, []string) {
	if doc.Slug == slug {
		return &doc, r.collectDocSlugs(doc)
	}
	for _, child := range doc.Children {
		if child.Slug == slug {
			return &child, r.collectDocSlugs(child)
		}
		for _, grandchild := range child.Children {
			if grandchild.Slug == slug {
				return &grandchild, r.collectDocSlugs(grandchild)
			}
		}
	}
	return nil, nil
}

// collectDocSlugs collects the slugs of the doc and its children.
func (r *docResource) collectDocSlugs(doc readme.CategoryDocs) []string {
	var slugs []string
	slugs = append(slugs, doc.Slug)
	for _, child := range doc.Children {
		slugs = append(slugs, r.collectDocSlugs(child)...)
	}
	return slugs
}

// handleChildDocs ensures that child docs are handled correctly before
// deleting the parent doc. If the provider configuration
// 'config.destroy_child_docs' is set to true, child docs will be deleted
// first. If child docs exist and 'config.destroy_child_docs' is false, an
// error will be returned.
func (r *docResource) handleChildDocs(
	ctx context.Context,
	resp *resource.DeleteResponse,
	slug string,
	parentDoc *readme.CategoryDocs,
) bool {
	if len(parentDoc.Children) > 0 && !r.config.DestroyChildDocs.ValueBool() {
		childSlugs := make([]string, 0, len(parentDoc.Children))
		for _, child := range parentDoc.Children {
			childSlugs = append(childSlugs, child.Slug)
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to delete doc '%s' because it has child docs.", slug),
			"Child docs must be deleted first. Set the provider 'config.destroy_child_docs' "+
				"option to true or delete child docs first.\n"+
				fmt.Sprintf("Docs that were found: %s", strings.Join(childSlugs, ", ")),
		)
		return false
	}
	return true
}

// deleteDocs deletes the docs in the slugsToDelete slice.
func (r *docResource) deleteDocs(
	ctx context.Context,
	resp *resource.DeleteResponse,
	slugsToDelete []string,
	requestOpts readme.RequestOptions,
) {
	for i := len(slugsToDelete) - 1; i >= 0; i-- {
		slug := slugsToDelete[i]
		if !r.deleteDoc(ctx, resp, slug, requestOpts) {
			return
		}
	}
}

// deleteDoc deletes the doc with the specified slug.
func (r *docResource) deleteDoc(
	ctx context.Context,
	resp *resource.DeleteResponse,
	slug string,
	requestOpts readme.RequestOptions,
) bool {
	_, apiResponse, err := r.client.Doc.Get(slug, requestOpts)
	if err != nil {
		if apiResponse != nil && apiResponse.HTTPResponse.StatusCode == 404 {
			tflog.Info(ctx, fmt.Sprintf("doc %s not found when deleting, removing from state", slug))
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Doc '%s' not found when deleting.", slug),
				"The doc was not found and has been removed from state.",
			)
			return false
		}
		resp.Diagnostics.AddError(
			"Unable to delete doc.",
			fmt.Sprintf("Error checking if doc '%s' exists: %s", slug, err.Error()),
		)
		return false
	}

	tflog.Info(ctx, fmt.Sprintf("deleting doc with slug %s and request options=%+v", slug, requestOpts))
	_, apiResponse, err = r.client.Doc.Delete(slug, requestOpts)
	if err != nil {
		if apiResponse != nil && apiResponse.HTTPResponse.StatusCode == 404 {
			tflog.Info(ctx, fmt.Sprintf("doc %s not found when deleting, removing from state", slug))
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Doc '%s' not found when deleting.", slug),
				"The doc was not found and has been removed from state.",
			)
			return false
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to delete doc with slug '%s'", slug),
			clientError(err, apiResponse),
		)
		return false
	}

	return true
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
		attrVal := IDPrefix + plan.ParentDoc.ValueString()
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
		Description:         docResourceDesc,
		MarkdownDescription: docResourceDesc,
		Attributes: map[string]schema.Attribute{
			"algolia": schema.SingleNestedAttribute{
				Description: "Metadata about the Algolia search integration. " +
					"See <https://docs.readme.com/main/docs/search> for more information.",
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"record_count": schema.Int64Attribute{
						Computed: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"publish_pending": schema.BoolAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
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
					"Accepts long page content, for example, greater than 100k characters. " +
					"Optionally use front matter to set certain attributes.",
				Computed: true,
				Optional: true,
			},
			"body_clean": schema.StringAttribute{
				Description: "The body content of the doc after transformations such as trimming leading and trailing" +
					"spaces.",
				Computed: true,
			},
			"body_html": schema.StringAttribute{
				Description: "The body content in HTML.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "**Required**. The category ID of the doc. Note that changing the category will result " +
					"in a replacement of the doc resource. Alternatively, set the `category` key the body front matter. " +
					"Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.",
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{},
				PlanModifiers: []planmodifier.String{
					frontmatter.GetString("Category"),
				},
			},
			"category_slug": schema.StringAttribute{
				Description: "**Required**. The category slug of the doc. Note that changing the category will result " +
					"in a replacement of the doc resource. Alternatively, set the `categorySlug` key the body front matter. " +
					"Docs that specify a `parent_doc` or `parent_doc_slug` will use their parent's category.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					frontmatter.GetString("CategorySlug"),
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
				Description: "Identifies if a doc is deprecated or not.",
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
					frontmatter.GetBool("Hidden"),
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
				Description: "Identifies if a doc is an API doc or not.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_reference": schema.BoolAttribute{
				Description: "Identifies if a doc is a reference doc or not.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"link_external": schema.BoolAttribute{
				Description: "Identifies a doc's link as external or not.",
				Computed:    true,
			},
			"link_url": schema.StringAttribute{
				Description: "The URL of the doc.",
				Computed:    true,
			},
			"metadata": schema.SingleNestedAttribute{
				Description: "Metadata about the doc.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "The description of the doc.",
						Computed:    true,
					},
					"image": schema.ListAttribute{
						Description: "An image associated with the doc.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"title": schema.StringAttribute{
						Description: "The title of the doc.",
						Computed:    true,
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
					frontmatter.GetInt64("Order"),
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
					frontmatter.GetString("ParentDoc"),
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
					frontmatter.GetString("ParentDocSlug"),
				},
			},
			"previous_slug": schema.StringAttribute{
				Description: "If the doc's slug has changed, this attribute contains the previous slug.",
				Computed:    true,
			},
			"project": schema.StringAttribute{
				Description: "The ID of the project the doc is in.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
					frontmatter.GetString("Title"),
				},
			},
			"type": schema.StringAttribute{
				Description: `**Required.** Type of the doc. The available types all show up under the /docs/ URL ` +
					`path of your docs project (also known as the "guides" section). Can be "basic" (most common), ` +
					`"error" (page describing an API error), or "link" (page that redirects to an external link).` +
					"This attribute may optionally be set in the body front matter.",
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					frontmatter.GetString("Type"),
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
			"use_slug": schema.StringAttribute{
				MarkdownDescription: "**Use with caution!** Create the doc resource by importing an existing doc by its slug. " +
					"This is non-conventional and should only be used when the slug is known and " +
					"the doc is not managed by Terraform or when the slug is changed in the web UI. " +
					"This is useful for managing an API specification's doc that gets created " +
					"automatically by ReadMe. When set, the specified doc will be replaced " +
					"with the Terraform-managed doc. " +
					"If this is set and then unset, a new doc will be created but the existing doc will not be " +
					"deleted. The existing doc will be orphaned and will not be managed by Terraform. " +
					"If this is unset and then set, the existing doc will be deleted and the resource will be " +
					"pointed to the specified doc. " +
					"In the case of API specification docs, the doc is implicitly deleted when the " +
					"API specification is deleted. " +
					"This attribute may be set in the body front matter with the `slug` key.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					frontmatter.GetString("Slug"),
				},
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
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}
