//nolint:goconst
package readme

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestProjectDataSource(t *testing.T) {
	defer gock.Off()
	gock.New(testURL).
		Get("/").
		Persist().
		Reply(200).
		JSON(map[string]string{
			"name":      "Test Project",
			"subdomain": "terraform-test.readme.io",
			"jwtSecret": "SuperSecret",
			"baseUrl":   "http://terraform-test.readme.io",
			"plan":      "enterprise",
		})

	tfConfig := `data "readme_project" "test" {}`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + tfConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the project response is returned with the data source.
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"name",
						"Test Project",
					),
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"subdomain",
						"terraform-test.readme.io",
					),
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"jwt_secret",
						"SuperSecret",
					),
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"base_url",
						"http://terraform-test.readme.io",
					),
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"plan",
						"enterprise",
					),
					// Verify placeholder id attribute.
					// See https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
					resource.TestCheckResourceAttr("data.readme_project.test", "id", "readme"),
				),
			},
		},
	})
}

func TestProjectDataSource_GetError(t *testing.T) {
	defer gock.Off()
	gock.New(testURL).
		Get("/").
		Persist().
		Reply(401).
		JSON(map[string]string{})

	expectError, _ := regexp.Compile(`Unable to retrieve project metadata\.`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `data "readme_project" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the project response is returned with the data source.
					resource.TestCheckResourceAttr(
						"data.readme_project.test",
						"name",
						"Test Project",
					),
				),

				ExpectError: expectError,
			},
		},
	})
}
