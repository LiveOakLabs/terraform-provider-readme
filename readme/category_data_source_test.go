// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

var mockCategory = readme.Category{
	CategoryType: "",
	CreatedAt:    "2023-01-06T21:25:39.543Z",
	ID:           "63b891d3ee384600680cea03",
	Order:        9999,
	Project:      "63b891d3ee384600680ce9eb",
	Reference:    false,
	Slug:         "documentation",
	Title:        "Documentation",
	Type:         "guide",
	Version:      mockVersion.ID,
}

func TestCategoryDataSource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Get("/categories/" + mockCategory.Slug).
						Persist().
						Reply(200).
						JSON(mockCategory)
					gock.New(testURL).
						Get("/version/1.1.1").
						Persist().
						Reply(200).
						JSON(mockVersion)
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
					gock.New(testURL).
						Get("/categories").
						MatchParam("perPage", "100").
						MatchParam("page", "1").
						Persist().
						Reply(200).
						AddHeader("link", `'<>; rel="next", <>; rel="prev", <>; rel="last"'`).
						AddHeader("x-total-count", "1").
						JSON(mockCategoryList)
				},
				Config: providerConfig + `data "readme_category" "test" { slug = "` + mockCategory.Slug + `" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"id",
						mockCategory.ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"created_at",
						mockCategory.CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"order",
						fmt.Sprintf("%v", mockCategory.Order),
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"project",
						mockCategory.Project,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"reference",
						strconv.FormatBool(mockCategory.Reference),
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"slug",
						mockCategory.Slug,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"title",
						mockCategory.Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"type",
						mockCategory.Type,
					),
					resource.TestCheckResourceAttr(
						"data.readme_category.test",
						"version_id",
						mockCategory.Version,
					),
				),
			},
		},
	})
}

func TestCategoryDataSource_GetError(t *testing.T) {
	expectError, _ := regexp.Compile(`Unable to retrieve category metadata`)
	expectResponse := readme.APIErrorResponse{
		Error:   "CATEGORY_NOTFOUND",
		Message: "The requested category could not be found.",
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
						Get("/categories/doesntexist").
						Times(1).
						Reply(404).
						JSON(expectResponse)
				},
				Config:      providerConfig + `data "readme_category" "test" { slug = "doesntexist" }`,
				ExpectError: expectError,
			},
		},
	})
}
