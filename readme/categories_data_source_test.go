package readme

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

var mockCategoryList = []readme.Category{mockCategory}

func TestCategoriesDataSource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
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
				Config: providerConfig + `data "readme_categories" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.id",
						mockCategoryList[0].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.created_at",
						mockCategoryList[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.order",
						fmt.Sprintf("%v", mockCategoryList[0].Order),
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.project",
						mockCategoryList[0].Project,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.reference",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.slug",
						mockCategoryList[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.title",
						mockCategoryList[0].Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.type",
						mockCategoryList[0].Type,
					),
					resource.TestCheckResourceAttr(
						"data.readme_categories.test",
						"categories.0.version_id",
						mockCategoryList[0].Version,
					),
				),
			},
		},
	})
}

func TestCategoriesDataSource_GetError(t *testing.T) {
	expectError, _ := regexp.Compile(`Unable to retrieve categories metadata`)

	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Get("/categories").
						Times(1).
						Reply(401).
						JSON(mockAPIError)
				},
				Config:      providerConfig + `data "readme_categories" "test" {}`,
				ExpectError: expectError,
			},
		},
	})
}
