// nolint:goconst // Intentional repetition of some values for tests.
package readme

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

// mockVersion represents a version response struct that's used throughout the
// version tests (version_data_source, versions_data_source, and version_resource).
var mockVersion = readme.Version{
	ID: "638cf4cfdea3ff0096d1a95a",
	Categories: []string{
		"638cf4cfdea3ff0096d1a95c",
	},
	Codename:     "A test version",
	CreatedAt:    "2022-12-04T19:28:15.190Z",
	ForkedFrom:   "63b891d3ee384600680ce9f9",
	IsBeta:       false,
	IsDeprecated: false,
	IsHidden:     false,
	IsStable:     true,
	Project:      "638cf4cedea3ff0096d1a955",
	ReleaseDate:  "2022-12-04T19:28:15.190Z",
	Version:      "1.1.1",
	VersionClean: "1.1.1",
}

// mockVersionList represents the list of version summaries used throughout tests.
var mockVersionList = []readme.VersionSummary{
	{
		Codename:     mockVersion.Codename,
		CreatedAt:    mockVersion.CreatedAt,
		ForkedFrom:   mockVersion.ForkedFrom,
		ID:           mockVersion.ID,
		IsBeta:       mockVersion.IsBeta,
		IsDeprecated: mockVersion.IsDeprecated,
		IsHidden:     mockVersion.IsHidden,
		IsStable:     mockVersion.IsStable,
		Version:      mockVersion.Version,
		VersionClean: mockVersion.VersionClean,
	},
}

func TestVersionDataSource(t *testing.T) {
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
						JSON(mockVersion)
				},
				Config: testProviderConfig + `data "readme_version" "test" {
					version_clean = "` + mockVersion.VersionClean + `"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"categories.0",
						mockVersion.Categories[0],
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"codename",
						mockVersion.Codename,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"created_at",
						mockVersion.CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"forked_from",
						mockVersion.ForkedFrom,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"id",
						mockVersion.ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_beta",
						strconv.FormatBool(mockVersion.IsBeta),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_deprecated",
						strconv.FormatBool(mockVersion.IsDeprecated),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_hidden",
						strconv.FormatBool(mockVersion.IsHidden),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_stable",
						strconv.FormatBool(mockVersion.IsStable),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"version",
						mockVersion.Version,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"version_clean",
						mockVersion.VersionClean,
					),
				),
			},
		},
	})
}

func TestVersionDataSource_GetByID(t *testing.T) {
	// Close all gocks when completed.
	defer gock.OffAll()
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
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
				},
				Config: testProviderConfig + `data "readme_version" "test" { id = "` + mockVersion.ID + `" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"categories.0",
						mockVersion.Categories[0],
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"codename",
						mockVersion.Codename,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"created_at",
						mockVersion.CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"id",
						mockVersion.ID,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_beta",
						strconv.FormatBool(mockVersion.IsBeta),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_deprecated",
						strconv.FormatBool(mockVersion.IsDeprecated),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_hidden",
						strconv.FormatBool(mockVersion.IsHidden),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"is_stable",
						strconv.FormatBool(mockVersion.IsStable),
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"version",
						mockVersion.Version,
					),
					resource.TestCheckResourceAttr(
						"data.readme_version.test",
						"version_clean",
						mockVersion.VersionClean,
					),
				),
			},
		},
	})
}

func TestVersionDataSource_AttributeError(t *testing.T) {
	expectError, _ := regexp.Compile(`'id', 'version', or 'version_clean' must be set\.`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testProviderConfig + `data "readme_version" "test" { }`,
				ExpectError: expectError,
			},
		},
	})
}

func TestVersionDataSource_GetError(t *testing.T) {
	expectError, _ := regexp.Compile(`Unable to read version`)

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
				Config:      testProviderConfig + `data "readme_version" "test" { version_clean = "1.1.1" }`,
				ExpectError: expectError,
			},
		},
	})
}
