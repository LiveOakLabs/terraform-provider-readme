package readme

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

// mockDoc is a ReadMe Doc type that's used by the data source and resource tests.
var mockDoc readme.Doc = readme.Doc{
	Algolia: readme.DocAlgolia{
		PublishPending: true,
		RecordCount:    0,
		UpdatedAt:      "2023-01-18T05:22:34.529Z",
	},
	API: readme.DocAPI{
		APISetting: "63c77e0bf76dee008e0c5197",
		Auth:       "required",
		Examples: readme.DocAPIExamples{
			Codes: []readme.DocAPIExamplesCodes{
				{
					Code:     "",
					Language: "go",
				},
			},
		},
		Method: "get",
		Params: []readme.DocAPIParams{
			{
				Name:       "example_path_param",
				Type:       "string",
				EnumValues: "",
				Default:    "foo",
				Desc:       "This is an example path param",
				Required:   false,
				In:         "path",
				Ref:        "",
				ID:         "63c77f40e790b60061a9a2ed",
			},
		},
		Results: readme.DocAPIResults{
			Codes: []readme.DocAPIResultsCodes{
				{
					Code:     `{"message": "ok"}`,
					Language: "json",
					Name:     "JSON",
					Status:   200,
				},
			},
		},
		URL: "/{something}",
	},
	Body:         "A turtle has been here.",
	BodyHTML:     "<div class=\"magic-block-textarea\"><p>A turtle has been here.</p></div>",
	Category:     mockCategory.ID,
	CreatedAt:    "2023-01-03T01:02:03.731Z",
	Deprecated:   false,
	Error:        readme.DocErrorObject{},
	Excerpt:      "",
	Hidden:       false,
	ID:           "63b37e8b65fd5b0057af23f1",
	Icon:         "",
	IsAPI:        false,
	IsReference:  false,
	LinkExternal: false,
	LinkURL:      "",
	Metadata: readme.DocMetadata{
		Description: "Example description.",
		Image: []any{
			"https://files.readme.io/53c37f9-logo.svg",
			"logo.svg",
			950,
			135,
			"#1c0e52",
		},
		Title: "A Test Doc",
	},
	Next: readme.DocNext{
		Description: "",
		Pages: []readme.DocNextPages{
			{
				Category:   "Documentation",
				Deprecated: false,
				Icon:       "file-text-o",
				Name:       "Another test doc",
				Slug:       "another-test-doc",
				Type:       "doc",
			},
		},
	},
	Order:         999,
	ParentDoc:     "633c5a54187d2c008e2e074c",
	PreviousSlug:  "",
	Project:       "638cf4cedea3ff0096d1a955",
	Revision:      2,
	Slug:          "a-test-doc",
	SlugUpdatedAt: "2023-01-02T01:44:37.530Z",
	SyncUnique:    "",
	Title:         "A Test Doc",
	Type:          "basic",
	UpdatedAt:     "2023-01-03T01:02:03.731Z",
	User:          "633c5a54187d2c008e2e074c",
	Version:       mockVersion.ID,
}

// makeMockDocParent returns a doc for use as a 'parent' doc.
func makeMockDocParent() readme.Doc {
	parentDoc := mockDoc
	parentDoc.Body = "This is a test parent doc."
	parentDoc.ID = mockDoc.ParentDoc
	parentDoc.ParentDoc = ""
	parentDoc.Slug = "test-parent-doc"
	parentDoc.Title = "Test Parent Doc"
	parentDoc.Hidden = false

	return parentDoc
}

var mockDocParent = makeMockDocParent()

// mockDocSearchResults represents the response from a doc search query.
var mockDocSearchResults = []readme.DocSearchResult{
	{
		IndexName:    "Page",
		InternalLink: "docs/" + mockDoc.Slug,
		IsReference:  false,
		LinkURL:      "",
		Method:       "get",
		ObjectID:     mockDoc.ID + "-1",
		Project:      "",
		ReferenceID:  mockDoc.ID,
		Slug:         mockDoc.Slug,
		Subdomain:    "",
		Title:        mockDoc.Title,
		Type:         mockDoc.Type,
		URL:          "",
		Version:      mockDoc.Version,
	},
	{
		IndexName:    "Page",
		InternalLink: "docs/" + mockDocParent.Slug,
		IsReference:  false,
		LinkURL:      "",
		Method:       "get",
		ObjectID:     mockDocParent.ID + "-1",
		Project:      mockDocParent.Project,
		ReferenceID:  mockDocParent.ID,
		Slug:         mockDocParent.Slug,
		Subdomain:    "",
		Title:        mockDocParent.Title,
		Type:         mockDocParent.Type,
		URL:          "",
		Version:      mockDocParent.Version,
	},
}

// mockDocSearchResponse represents the API response to a search query.
var mockDocSearchResponse = readme.DocSearchResults{
	Results: mockDocSearchResults,
}

// docCommonGocks provides Gocks that are commonly used throughout the doc tests.
// This includes the category, version, and parent doc lookup responses.
func docCommonGocks() {
	// Get category list to lookup category.
	gock.New(testURL).
		Get("/categories").
		MatchParam("perPage", "100").
		MatchParam("page", "1").
		Persist().
		Reply(200).
		AddHeader("link", `'<>; rel="next", <>; rel="prev", <>; rel="last"'`).
		AddHeader("x-total-count", "1").
		JSON(mockCategoryList)
	// Lookup category to get slug.
	gock.New(testURL).
		Get("/categories/" + mockCategory.Slug).
		Persist().
		Reply(200).
		JSON(mockCategory)
	// Lookup version.
	gock.New(testURL).Get("/version").Persist().Reply(200).JSON(mockVersionList)
	// List of docs to match parent doc.
	gock.New(testURL).
		Post("/docs").
		Path("search").
		MatchParam("search", mockDocParent.ID).
		Persist().
		Reply(200).
		JSON(mockDocSearchResponse)
	// Lookup parent doc.
	gock.New(testURL).Get("/docs/" + mockDocParent.Slug).Persist().Reply(200).JSON(mockDocParent)
}

// docResourceCommonChecks returns all attribute checks for the data source and resource.
func docResourceCommonChecks(mock readme.Doc, prefix string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.api_setting",
			mock.API.APISetting,
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "api.auth", mock.API.Auth),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.examples.codes.0.code",
			mock.API.Examples.Codes[0].Code,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.examples.codes.0.language",
			mock.API.Examples.Codes[0].Language,
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "api.method", mock.API.Method),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.default",
			mock.API.Params[0].Default,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.desc",
			mock.API.Params[0].Desc,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.enum_values",
			mock.API.Params[0].EnumValues,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.id",
			mock.API.Params[0].ID,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.in",
			mock.API.Params[0].In,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.name",
			mock.API.Params[0].Name,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.ref",
			mock.API.Params[0].Ref,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.required",
			fmt.Sprintf("%v", mock.API.Params[0].Required),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.params.0.type",
			mock.API.Params[0].Type,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.results.codes.0.code",
			mock.API.Results.Codes[0].Code,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.results.codes.0.language",
			mock.API.Results.Codes[0].Language,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.results.codes.0.name",
			mock.API.Results.Codes[0].Name,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"api.results.codes.0.status",
			fmt.Sprintf("%v", mock.API.Results.Codes[0].Status),
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "api.url", mock.API.URL),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "body", mock.Body),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "body_html", mock.BodyHTML),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "category", mock.Category),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "created_at", mock.CreatedAt),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"deprecated",
			fmt.Sprintf("%v", mock.Deprecated),
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "excerpt", mock.Excerpt),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"hidden",
			fmt.Sprintf("%v", mock.Hidden),
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "icon", mock.Icon),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "id", mock.ID),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"is_api",
			fmt.Sprintf("%v", mock.IsAPI),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"is_reference",
			fmt.Sprintf("%v", mock.IsReference),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"link_external",
			fmt.Sprintf("%v", mock.LinkExternal),
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "link_url", mock.LinkURL),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"metadata.description",
			mock.Metadata.Description,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"metadata.image.0",
			mock.Metadata.Image[0].(string),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"metadata.title",
			mock.Metadata.Title,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.category",
			mock.Next.Pages[0].Category,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.deprecated",
			fmt.Sprintf("%v", mock.Next.Pages[0].Deprecated),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.icon",
			mock.Next.Pages[0].Icon,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.name",
			mock.Next.Pages[0].Name,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.slug",
			mock.Next.Pages[0].Slug,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"next.pages.0.type",
			mock.Next.Pages[0].Type,
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"order",
			fmt.Sprintf("%v", mock.Order),
		),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"previous_slug",
			mock.PreviousSlug,
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "project", mock.Project),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"revision",
			fmt.Sprintf("%v", mock.Revision),
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "slug", mock.Slug),
		resource.TestCheckResourceAttr(
			prefix+"readme_doc.test",
			"slug_updated_at",
			mock.SlugUpdatedAt,
		),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "sync_unique", mock.SyncUnique),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "title", mock.Title),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "type", mock.Type),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "updated_at", mock.UpdatedAt),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "user", mock.User),
		resource.TestCheckResourceAttr(prefix+"readme_doc.test", "version_id", mock.Version),
	)
}
