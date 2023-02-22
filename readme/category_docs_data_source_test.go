package readme

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

func TestCategoryDocsDataSource(t *testing.T) {
	expectResponse := []readme.CategoryDocs{
		{
			ID:     "63b891d3ee384600680cea15",
			Title:  "Getting Started",
			Slug:   "getting-started",
			Order:  999,
			Hidden: false,
			Children: []readme.CategoryDocsChildren{
				{
					ID:     "63bdfb0079110f0094641789",
					Title:  "Test Child Doc",
					Slug:   "test-child-doc",
					Order:  999,
					Hidden: false,
				},
			},
		},
	}

	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Get("/categories/testing/docs").
						Persist().
						Reply(200).
						JSON(expectResponse)
				},
				Config: providerConfig + `data "readme_category_docs" "test" { slug = "testing" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.id",
						"63b891d3ee384600680cea15",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.title",
						"Getting Started",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.slug",
						"getting-started",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.order",
						"999",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.hidden",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.children.0.id",
						"63bdfb0079110f0094641789",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.children.0.slug",
						"test-child-doc",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.children.0.title",
						"Test Child Doc",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.children.0.order",
						"999",
					),
					resource.TestCheckResourceAttr(
						"data.readme_category_docs.test",
						"docs.0.children.0.hidden",
						"false",
					),
				),
			},
		},
	})
}

func TestCategoryDocsDataSource_GetError(t *testing.T) {
	expectResponse := readme.APIErrorResponse{
		Error:   "CATEGORY_NOTFOUND",
		Message: "The requested category could not be found.",
	}
	expectError, _ := regexp.Compile(`Unable to retrieve category docs`)

	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Get("/categories/testing/docs").
						Times(1).
						Reply(404).
						JSON(expectResponse)
				},
				Config:      providerConfig + `data "readme_category_docs" "test" { slug = "testing" }`,
				ExpectError: expectError,
			},
		},
	})
}
