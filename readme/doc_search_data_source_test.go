// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestDocSearchDataSource(t *testing.T) {
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
						Post("/docs").
						Path("search").
						Persist().
						Reply(200).
						JSON(mockDocSearchResponse)
				},
				Config: providerConfig + `
					data "readme_doc_search" "test" {
						query = "*"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.index_name",
						mockDocSearchResults[0].IndexName,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.internal_link",
						mockDocSearchResults[0].InternalLink,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.is_reference",
						fmt.Sprintf("%v", mockDocSearchResults[0].IsReference),
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.link_url",
						mockDocSearchResults[0].LinkURL,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.method",
						mockDocSearchResults[0].Method,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.object_id",
						mockDocSearchResults[0].ObjectID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.project",
						mockDocSearchResults[0].Project,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.reference_id",
						mockDocSearchResults[0].ReferenceID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.slug",
						mockDocSearchResults[0].Slug,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.subdomain",
						mockDocSearchResults[0].Subdomain,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.title",
						mockDocSearchResults[0].Title,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.type",
						mockDocSearchResults[0].Type,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.url",
						mockDocSearchResults[0].URL,
					),
					resource.TestCheckResourceAttr(
						"data.readme_doc_search.test",
						"results.0.version",
						mockDocSearchResults[0].Version,
					),
				),
			},
		},
	})
}
