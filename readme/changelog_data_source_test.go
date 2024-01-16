// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestChangelogDataSource(t *testing.T) {
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
						Get("/changelogs/" + mockChangelogs[0].Slug).
						Persist().
						Reply(200).
						JSON(mockChangelogs[0])
				},
				Config: providerConfig + `
					data "readme_changelog" "test" {
						slug = "` + mockChangelogs[0].Slug + `"
					}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"id",
						mockChangelogs[0].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"title",
						mockChangelogs[0].Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"slug",
						mockChangelogs[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"body",
						mockChangelogs[0].Body,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"created_at",
						mockChangelogs[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"updated_at",
						mockChangelogs[0].UpdatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"html",
						mockChangelogs[0].HTML,
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"revision",
						fmt.Sprintf("%v", mockChangelogs[0].Revision),
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"metadata.image.0",
						fmt.Sprintf("%v", mockChangelogs[0].Metadata.Image[0]),
					),
					resource.TestCheckResourceAttr(
						"data.readme_changelog.test",
						"algolia.updated_at",
						mockChangelogs[0].Algolia.UpdatedAt,
					),
				),
			},
		},
	})
}
