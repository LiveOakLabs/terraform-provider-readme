package readme

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestCustomPagesDataSource(t *testing.T) {
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
						Get("/custompages").
						MatchParam("perPage", "100").
						MatchParam("page", "1").
						Persist().
						Reply(200).
						AddHeader("link", `'<>; rel="next", <>; rel="prev", <>; rel="last"'`).
						AddHeader("x-total-count", "1").
						JSON(mockCustomPages)
				},
				Config: providerConfig + `data "readme_custom_pages" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_custom_pages.test",
						"results.0.id",
						mockCustomPages[0].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_pages.test",
						"results.0.title",
						mockCustomPages[0].Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_pages.test",
						"results.1.id",
						mockCustomPages[1].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_custom_pages.test",
						"results.1.title",
						mockCustomPages[1].Title,
					),
				),
			},
		},
	})
}
