// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestCustomPageDataSource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.OffAll()
					gock.New(testURL).
						Get("/custompages/" + mockCustomPages[0].Slug).
						Persist().
						Reply(200).
						JSON(mockCustomPages[0])
				},
				Config: providerConfig + `
					data "readme_custom_page" "test" {
						slug = "` + mockCustomPages[0].Slug + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"id",
						mockCustomPages[0].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"title",
						mockCustomPages[0].Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"slug",
						mockCustomPages[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"body",
						mockCustomPages[0].Body,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"created_at",
						mockCustomPages[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"updated_at",
						mockCustomPages[0].UpdatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"fullscreen",
						fmt.Sprintf("%v", mockCustomPages[0].Fullscreen),
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"html",
						mockCustomPages[0].HTML,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"htmlmode",
						fmt.Sprintf("%v", mockCustomPages[0].HTMLMode),
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"revision",
						fmt.Sprintf("%v", mockCustomPages[0].Revision),
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"metadata.image.0",
						fmt.Sprintf("%v", mockCustomPages[0].Metadata.Image[0]),
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_page.test",
						"algolia.updated_at",
						mockCustomPages[0].Algolia.UpdatedAt,
					),
				),
			},
		},
	})
}
