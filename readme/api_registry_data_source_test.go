// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestAPIRegistryDataSource(t *testing.T) {
	gock.New(testURL).
		Get("/").
		Persist().
		Reply(200).
		JSON(`{"one": "two"}`)
	defer gock.Off()

	tfConfig := `data "readme_api_registry" "test" { uuid = "somethingUnique" }`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + tfConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_api_registry.test",
						"uuid",
						"somethingUnique",
					),
					resource.TestCheckResourceAttr("data.readme_api_registry.test", "id", "readme"),
					resource.TestCheckResourceAttr(
						"data.readme_api_registry.test",
						"definition",
						`{"one": "two"}`,
					),
				),
			},
		},
	})
}

func TestAPIRegistryDataSource_GetError(t *testing.T) {
	gock.New(testURL).
		Get("/").
		Persist().
		Reply(401).
		JSON(map[string]string{})
	defer gock.Off()

	expectError, _ := regexp.Compile(
		`Unable to retrieve API registry metadata\.`,
	)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "readme_api_registry" "test" { uuid = "somethingUnique" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_api_registry.test",
						"uuid",
						"somethingUnique",
					),
				),

				ExpectError: expectError,
			},
		},
	})
}
