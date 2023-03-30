package readme

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// docModel defines the fields and their types that map to the schemas.
type docModel struct {
	Algolia         types.Object `tfsdk:"algolia"`
	API             types.Object `tfsdk:"api"`
	Body            types.String `tfsdk:"body"`
	BodyClean       types.String `tfsdk:"body_clean"`
	BodyHTML        types.String `tfsdk:"body_html"`
	Category        types.String `tfsdk:"category"`
	CategorySlug    types.String `tfsdk:"category_slug"`
	CreatedAt       types.String `tfsdk:"created_at"`
	Deprecated      types.Bool   `tfsdk:"deprecated"`
	Excerpt         types.String `tfsdk:"excerpt"`
	Hidden          types.Bool   `tfsdk:"hidden"`
	ID              types.String `tfsdk:"id"`
	Icon            types.String `tfsdk:"icon"`
	IsAPI           types.Bool   `tfsdk:"is_api"`
	IsReference     types.Bool   `tfsdk:"is_reference"`
	LinkExternal    types.Bool   `tfsdk:"link_external"`
	LinkURL         types.String `tfsdk:"link_url"`
	Error           types.Object `tfsdk:"error"`
	Metadata        types.Object `tfsdk:"metadata"`
	Next            types.Object `tfsdk:"next"`
	ParentDoc       types.String `tfsdk:"parent_doc"`
	ParentDocSlug   types.String `tfsdk:"parent_doc_slug"`
	Order           types.Int64  `tfsdk:"order"`
	PreviousSlug    types.String `tfsdk:"previous_slug"`
	Project         types.String `tfsdk:"project"`
	Revision        types.Int64  `tfsdk:"revision"`
	Slug            types.String `tfsdk:"slug"`
	SlugUpdatedAt   types.String `tfsdk:"slug_updated_at"`
	SyncUnique      types.String `tfsdk:"sync_unique"`
	Title           types.String `tfsdk:"title"`
	Type            types.String `tfsdk:"type"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	User            types.String `tfsdk:"user"`
	VerifyParentDoc types.Bool   `tfsdk:"verify_parent_doc"`
	Version         types.String `tfsdk:"version"`
	VersionID       types.String `tfsdk:"version_id"`
}

// docModelValue returns a docModel value with the fields mapped from the `doc` parameter.
//
// This is used by both the data source and resource to create a plan and state value.
//
// The `model` parameter represents a `docModel` that is merged with the response model.
// This should include fields/attributes that are not in the API response, such as the
// "slug" fields and "version" field.
func docModelValue(ctx context.Context, doc readme.Doc, model docModel) docModel {
	// Ensure ParentDocSlug has a value. This is an optional value that is
	// only otherwise resolved if a parent doc is set.
	if model.ParentDocSlug.IsUnknown() {
		model.ParentDocSlug = types.StringValue("")
	}

	return docModel{
		Algolia:         docModelAlgoliaValue(doc.Algolia),
		API:             docModelAPIValue(doc.API),
		Body:            model.Body,
		BodyClean:       types.StringValue(doc.Body),
		BodyHTML:        types.StringValue(doc.BodyHTML),
		Category:        types.StringValue(doc.Category),
		CategorySlug:    model.CategorySlug,
		CreatedAt:       types.StringValue(doc.CreatedAt),
		Deprecated:      types.BoolValue(doc.Deprecated),
		Error:           docModelErrorValue(doc.Error),
		Excerpt:         types.StringValue(doc.Excerpt),
		Hidden:          types.BoolValue(doc.Hidden),
		ID:              types.StringValue(doc.ID),
		Icon:            types.StringValue(doc.Icon),
		IsAPI:           types.BoolValue(doc.IsAPI),
		IsReference:     types.BoolValue(doc.IsReference),
		LinkExternal:    types.BoolValue(doc.LinkExternal),
		LinkURL:         types.StringValue(doc.LinkURL),
		Metadata:        docModelMetadataValue(doc.Metadata),
		Next:            docModelNextValue(doc.Next),
		Order:           types.Int64Value(int64(doc.Order)),
		ParentDoc:       types.StringValue(doc.ParentDoc),
		ParentDocSlug:   model.ParentDocSlug,
		PreviousSlug:    types.StringValue(doc.PreviousSlug),
		Project:         types.StringValue(doc.Project),
		Revision:        types.Int64Value(int64(doc.Revision)),
		Slug:            types.StringValue(doc.Slug),
		SlugUpdatedAt:   types.StringValue(doc.SlugUpdatedAt),
		SyncUnique:      types.StringValue(doc.SyncUnique),
		Title:           types.StringValue(doc.Title),
		Type:            types.StringValue(doc.Type),
		UpdatedAt:       types.StringValue(doc.UpdatedAt),
		User:            types.StringValue(doc.User),
		VerifyParentDoc: model.VerifyParentDoc,
		Version:         model.Version,
		VersionID:       types.StringValue(doc.Version),
	}
}

// getDoc retrieves a doc and returns the Terraform data source or resource.
//
// The `model` parameter represents a `docModel` that is merged with the response model.
// This should include fields/attributes that are not in the API response, such as the
// "slug" fields and "version" field.
//
// It returns a `docModel` for use in a plan or state.
func getDoc(
	client *readme.Client,
	ctx context.Context,
	slug string,
	model docModel,
	options readme.RequestOptions,
) (docModel, *readme.APIResponse, error) {
	var state docModel

	// Get the doc from ReadMe.
	response, apiResponse, err := client.Doc.Get(slug, options)
	if err != nil {
		return state, apiResponse, fmt.Errorf(clientError(err, apiResponse))
	}

	// Map the API object to the Terraform model.
	state = docModelValue(ctx, response, model)

	// Resolve the 'version' attribute if it's not set.
	if state.Version.ValueString() == "" && state.VersionID.ValueString() != "" {
		tflog.Info(
			ctx,
			fmt.Sprintf("resolving version for version_id %s", state.VersionID.ValueString()),
		)
		state.Version = types.StringValue(
			versionClean(ctx, client, state.VersionID.ValueString()),
		)
	}

	// Resolve the 'category_slug' attribute if it's not set.
	if state.CategorySlug.ValueString() == "" {
		tflog.Info(
			ctx,
			fmt.Sprintf("resolving category_slug for category %s", state.Category.ValueString()),
		)

		category, apiResponse, err := client.Category.Get(
			"id:"+state.Category.ValueString(),
			options,
		)
		if err != nil {
			return state, apiResponse, errors.New(clientError(err, apiResponse))
		}
		state.CategorySlug = types.StringValue(category.Slug)
	}

	// Resolve the 'parent_doc_slug' attribute if 'parent_doc' is set.
	if state.VerifyParentDoc.IsNull() || state.VerifyParentDoc.ValueBool() {
		if state.ParentDoc.ValueString() != "" && state.ParentDocSlug.ValueString() == "" {
			tflog.Info(
				ctx,
				fmt.Sprintf(
					"resolving parent_doc_slug for parent_doc %s",
					state.ParentDoc.ValueString(),
				),
			)

			parent, apiResponse, err := client.Doc.Get("id:"+state.ParentDoc.ValueString(), options)
			if err != nil {
				// failing here
				return state, apiResponse, errors.New(clientError(err, apiResponse))
			}
			state.ParentDocSlug = types.StringValue(parent.Slug)
		}
	}

	return state, apiResponse, nil
}

// docModelAlgoliaValue returns the populated `algolia` object value embedded within `docModel`.
func docModelAlgoliaValue(algolia readme.DocAlgolia) basetypes.ObjectValue {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"record_count":    types.Int64Type,
			"publish_pending": types.BoolType,
			"updated_at":      types.StringType,
		},
		map[string]attr.Value{
			"record_count":    types.Int64Value(int64(algolia.RecordCount)),
			"publish_pending": types.BoolValue(algolia.PublishPending),
			"updated_at":      types.StringValue(algolia.UpdatedAt),
		},
	)
}

// docModelAPIParamsTypes is the attribute types map for the `api:params` object embedded within `docModel`.
var docModelAPIParamsTypes = map[string]attr.Type{
	"name":        types.StringType,
	"type":        types.StringType,
	"enum_values": types.StringType,
	"default":     types.StringType,
	"desc":        types.StringType,
	"in":          types.StringType,
	"ref":         types.StringType,
	"id":          types.StringType,
	"required":    types.BoolType,
}

// docModelAPIExamplesCodesTypes is the attribute types map for the `api:examples:codes` object.
var docModelAPIExamplesCodesTypes = map[string]attr.Type{
	"code":     types.StringType,
	"language": types.StringType,
}

// docModelAPIResultsCodesTypes is the attribute types map for the `api:results:codes` object.
var docModelAPIResultsCodesTypes = map[string]attr.Type{
	"code":     types.StringType,
	"language": types.StringType,
	"name":     types.StringType,
	"status":   types.Int64Type,
}

// docModelAPIValue returns the populated `api` object value embedded within `docModel`.
func docModelAPIValue(api readme.DocAPI) basetypes.ObjectValue {
	// map of `api:examples` attribute types.
	docModelMapAPIExamples := map[string]attr.Type{
		"codes": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: docModelAPIExamplesCodesTypes,
			},
		},
	}

	// map of `api:results` attribute types.
	docModelMapAPIResults := map[string]attr.Type{
		"codes": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: docModelAPIResultsCodesTypes,
			},
		},
	}

	// return the nested `api` attribute value.
	return types.ObjectValueMust(
		// Map of the attribute types.
		map[string]attr.Type{
			"api_setting": types.StringType,
			"auth":        types.StringType,
			"method":      types.StringType,
			"url":         types.StringType,
			"examples":    types.ObjectType{AttrTypes: docModelMapAPIExamples},
			"params": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: docModelAPIParamsTypes,
				},
			},
			"results": types.ObjectType{AttrTypes: docModelMapAPIResults},
		},
		// Map of the attribute values.
		map[string]attr.Value{
			"api_setting": types.StringValue(api.APISetting),
			"auth":        types.StringValue(api.Auth),
			"method":      types.StringValue(api.Method),
			"url":         types.StringValue(api.URL),
			"examples": types.ObjectValueMust(
				docModelMapAPIExamples,
				map[string]attr.Value{
					"codes": types.ListValueMust(
						types.ObjectType{
							AttrTypes: docModelAPIExamplesCodesTypes,
						},
						docModelAPIExamplesCodesValue(api),
					),
				},
			),
			"params": types.ListValueMust(
				types.ObjectType{
					AttrTypes: docModelAPIParamsTypes,
				},
				docModelAPIParamsValue(api),
			),
			"results": types.ObjectValueMust(
				docModelMapAPIResults,
				map[string]attr.Value{
					"codes": types.ListValueMust(
						types.ObjectType{
							AttrTypes: docModelAPIResultsCodesTypes,
						},
						docModelAPIResultsCodesValue(api),
					),
				},
			),
		},
	)
}

// docModelAPIExamplesCodesValue returns the populated `api:examples:codes` object value.
func docModelAPIExamplesCodesValue(api readme.DocAPI) []attr.Value {
	codes := []attr.Value{}
	for _, code := range api.Examples.Codes {
		codes = append(codes, types.ObjectValueMust(
			docModelAPIExamplesCodesTypes,
			map[string]attr.Value{
				"code":     types.StringValue(code.Code),
				"language": types.StringValue(code.Language),
			},
		))
	}

	return codes
}

// docModelApiExamplesCodesValue returns the populated `api:examples:codes` object value.
func docModelAPIResultsCodesValue(api readme.DocAPI) []attr.Value {
	codes := []attr.Value{}
	for _, code := range api.Results.Codes {
		codes = append(codes, types.ObjectValueMust(
			docModelAPIResultsCodesTypes,
			map[string]attr.Value{
				"code":     types.StringValue(code.Code),
				"language": types.StringValue(code.Language),
				"name":     types.StringValue(code.Name),
				"status":   types.Int64Value(int64(code.Status)),
			},
		))
	}

	return codes
}

// docModelAPIParamsValue returns the populated `api:params` object values.
func docModelAPIParamsValue(api readme.DocAPI) []attr.Value {
	params := []attr.Value{}
	for _, param := range api.Params {
		params = append(params, types.ObjectValueMust(
			docModelAPIParamsTypes,
			map[string]attr.Value{
				"name":        types.StringValue(param.Name),
				"type":        types.StringValue(param.Type),
				"enum_values": types.StringValue(param.EnumValues),
				"default":     types.StringValue(param.Default),
				"desc":        types.StringValue(param.Desc),
				"in":          types.StringValue(param.In),
				"ref":         types.StringValue(param.Ref),
				"id":          types.StringValue(param.ID),
				"required":    types.BoolValue(param.Required),
			},
		))
	}

	return params
}

// docModelErrorValue returns the populated `error` object value embedded within `docModel`.
func docModelErrorValue(docErr readme.DocErrorObject) basetypes.ObjectValue {
	// Return a null object if the error key is empty.
	if docErr.Code == "" {
		return types.ObjectNull(
			map[string]attr.Type{
				"code": types.StringType,
			},
		)
	}

	// Return the populated error object value.
	return types.ObjectValueMust(
		map[string]attr.Type{
			"code": types.StringType,
		},
		map[string]attr.Value{
			"code": types.StringValue(docErr.Code),
		},
	)
}

// docModelMetadataValue returns the populated `metadata` object value embedded within `docModel`.
func docModelMetadataValue(metadata readme.DocMetadata) basetypes.ObjectValue {
	// metadataTypes is the map of Terraform attribute types for the `metadata` key.
	metadataTypes := map[string]attr.Type{
		"title":       types.StringType,
		"description": types.StringType,
		"image": types.ListType{
			ElemType: types.StringType,
		},
	}

	// Build list of images.
	images := []attr.Value{}
	for _, img := range metadata.Image {
		images = append(images, types.StringValue(fmt.Sprintf("%v", img)))
	}

	// Return a null object if the metadata is empty.
	if metadata.Description == "" {
		return types.ObjectNull(metadataTypes)
	}

	// Return the populated metadata object value.
	return types.ObjectValueMust(
		metadataTypes,
		map[string]attr.Value{
			"title":       types.StringValue(metadata.Title),
			"description": types.StringValue(metadata.Description),
			"image":       types.ListValueMust(types.StringType, images),
		},
	)
}

// docModelNextValue returns the populated `next` object value embedded within `docModel`.
func docModelNextValue(next readme.DocNext) basetypes.ObjectValue {
	// pagesTypes is the map of Terraform attribute types for the `next:pages` key.
	pagesTypes := map[string]attr.Type{
		"category":   types.StringType,
		"deprecated": types.BoolType,
		"icon":       types.StringType,
		"name":       types.StringType,
		"slug":       types.StringType,
		"type":       types.StringType,
	}

	// Build list of pages.
	pages := []attr.Value{}
	for _, page := range next.Pages {
		pages = append(pages, types.ObjectValueMust(
			pagesTypes,
			map[string]attr.Value{
				"category":   types.StringValue(page.Category),
				"deprecated": types.BoolValue(page.Deprecated),
				"icon":       types.StringValue(page.Icon),
				"name":       types.StringValue(page.Name),
				"slug":       types.StringValue(page.Slug),
				"type":       types.StringValue(page.Type),
			},
		))
	}

	// Return populated 'next' object value.
	return types.ObjectValueMust(
		map[string]attr.Type{
			"description": types.StringType,
			"pages": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: pagesTypes,
				},
			},
		},
		map[string]attr.Value{
			"description": types.StringValue(next.Description),
			"pages": types.ListValueMust(
				types.ObjectType{
					AttrTypes: pagesTypes,
				},
				pages,
			),
		},
	)
}
