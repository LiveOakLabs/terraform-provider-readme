package readme

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gopkg.in/h2non/gock.v1"
)

func TestVersionsDataSource(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Get("/version").
						Persist().
						Reply(200).
						JSON(mockVersionList)
				},
				Config: testProviderConfig + `data "readme_versions" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.id",
						mockVersionList[0].ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.codename",
						mockVersionList[0].Codename,
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.created_at",
						mockVersionList[0].CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.is_beta",
						strconv.FormatBool(mockVersionList[0].IsBeta),
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.is_deprecated",
						strconv.FormatBool(mockVersionList[0].IsDeprecated),
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.is_hidden",
						strconv.FormatBool(mockVersionList[0].IsHidden),
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.is_stable",
						strconv.FormatBool(mockVersionList[0].IsStable),
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.version",
						mockVersionList[0].Version,
					),
					resource.TestCheckResourceAttr(
						"data.readme_versions.test",
						"versions.0.version_clean",
						mockVersionList[0].VersionClean,
					),
				),
			},
		},
	})
}

func TestVersionsDataSource_GetError(t *testing.T) {
	expectError, _ := regexp.Compile(`Unable to read versions`)

	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).Get("/version").Times(1).Reply(401).JSON(map[string]string{})
				},
				Config:      testProviderConfig + `data "readme_versions" "test" {}`,
				ExpectError: expectError,
			},
		},
	})
}
