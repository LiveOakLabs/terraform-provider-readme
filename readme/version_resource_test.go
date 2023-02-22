package readme

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/lobliveoaklabs/readme-api-go-client/readme"
	"gopkg.in/h2non/gock.v1"
)

func TestVersionResource(t *testing.T) {
	// mockUpdatedVersion is used in the update tests.
	mockUpdatedVersion := mockVersion
	mockUpdatedVersion.Codename = "Example Updated"
	mockUpdatedVersion.Version = "1.1.2"
	mockUpdatedVersion.VersionClean = "1.1.2"

	// Close all gocks after completion.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test creating a version.
			{
				Config: providerConfig + `resource "readme_version" "test" {
					from      = "1.0.0"
					version   = "` + mockVersion.Version + `"
					codename  = "` + mockVersion.Codename + `"
					is_stable = ` + strconv.FormatBool(mockVersion.IsStable) + `
				}`,
				PreConfig: func() {
					// Post-create read and refresh.
					gock.New(testURL).
						Get("/version/" + mockVersion.Version).
						Times(2).
						Reply(200).
						JSON(mockVersion)
					// Create resource.
					gock.New(testURL).
						Post("/version").
						Times(1).
						Reply(200).
						JSON(mockVersion)
					gock.New(testURL).Delete("/version/" + mockVersion.Version).Times(5).Reply(200)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"version",
						mockVersion.Version,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"codename",
						mockVersion.Codename,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"created_at",
						mockVersion.CreatedAt,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"forked_from",
						mockVersion.ForkedFrom,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"is_beta",
						strconv.FormatBool(mockVersion.IsBeta),
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"is_deprecated",
						strconv.FormatBool(mockVersion.IsDeprecated),
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"is_hidden",
						strconv.FormatBool(mockVersion.IsHidden),
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"is_stable",
						strconv.FormatBool(mockVersion.IsStable),
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"project",
						mockVersion.Project,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"release_date",
						mockVersion.ReleaseDate,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"version",
						mockVersion.Version,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"version_clean",
						mockVersion.VersionClean,
					),
				),
			},
			// Test updating the codename.
			{
				Config: providerConfig + `resource "readme_version" "test" {
					from     = "1.0.0"
					version  = "` + mockUpdatedVersion.Version + `"
					codename = "` + mockUpdatedVersion.Codename + `"
				}`,
				PreConfig: func() {
					// Post-update read.
					gock.New(testURL).
						Get("/version/" + mockUpdatedVersion.Version).
						Times(2).
						Reply(200).
						JSON(mockUpdatedVersion)
					// Update resource.
					gock.New(testURL).
						Put("/version/" + mockVersion.Version).
						Times(1).
						Reply(200).
						JSON(mockUpdatedVersion)
					// Read current version.
					gock.New(testURL).
						Get("/version/" + mockVersion.Version).
						Times(2).
						Reply(200).
						JSON(mockUpdatedVersion)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"version",
						mockUpdatedVersion.Version,
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"from",
						"1.0.0",
					),
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"codename",
						mockUpdatedVersion.Codename,
					),
				),
			},
			// Test importing.
			{
				ResourceName:      "readme_version.test",
				ImportState:       true,
				ImportStateId:     mockUpdatedVersion.Version,
				ImportStateVerify: true,
				// The 'from' attribute is tracked internally and is not returned from the API, therefore there is no
				// value for it during import.
				ImportStateVerifyIgnore: []string{"from"},
			},
			// When is_stable gets updated, the resource will be re-created.
			{
				Config: providerConfig + `resource "readme_version" "test" {
						from     = "1.0.0"
						version  = "` + mockUpdatedVersion.Version + `"
						codename = "` + mockUpdatedVersion.Codename + `"
						is_stable = false
					}`,
				PreConfig: func() {
					mockUpdatedVersion.IsStable = false
					gock.OffAll()
					// Expect the provider to read the updated version twice.
					gock.New(testURL).
						Get("/version/" + mockUpdatedVersion.Version).
						Times(2).
						Reply(200).
						JSON(mockUpdatedVersion)
					// Expect the replacement to POST a new version.
					gock.New(testURL).
						Post("/version").
						Times(1).
						Reply(200).
						JSON(mockUpdatedVersion)
					// Expect the current version to be deleted for replacement.
					gock.New(testURL).
						Delete("/version/" + mockUpdatedVersion.Version).
						Times(1).
						Reply(200)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"readme_version.test",
						"version",
						mockUpdatedVersion.Version,
					),
					resource.TestCheckResourceAttr("readme_version.test", "is_stable", "false"),
				),
			},
		},
	})
}

func TestVersionsResource_Error(t *testing.T) {
	// expectCreateResponse is what is expected upon creation.
	expectCreateResponse := readme.APIErrorResponse{
		Error:      "VERSION_DUPLICATE",
		Message:    "The version already exists.",
		Suggestion: "",
		Docs:       "",
		Help:       "",
		Poem:       []string{"one"},
	}

	expectError, _ := regexp.Compile(`Unable to create version`)

	// Close all gocks when completed.
	defer gock.OffAll()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					gock.New(testURL).
						Post("/version").
						Times(1).
						Reply(400).
						JSON(expectCreateResponse)
				},
				Config: providerConfig + `resource "readme_version" "test" {
					from    = "1.0.0"
					version = "` + mockVersion.Version + `"
					codename = "` + mockVersion.Codename + `"
				}`,
				ExpectError: expectError,
			},
		},
	})
}

func TestVersionsResource_AttributeError(t *testing.T) {
	testCases := []struct {
		ExpectError string
		Config      string
	}{
		{
			ExpectError: "A stable version cannot be hidden",
			Config: `resource "readme_version" "test" {
				from    = "1.1.0"
				version = "1.1.0"
				codename = "Example"
				is_stable = true
				is_hidden = true
			}
			`,
		},
		{
			ExpectError: "A stable version cannot be deprecated",
			Config: `resource "readme_version" "test" {
				from    = "1.1.0"
				version = "1.1.0"
				codename = "Example"
				is_stable = true
				is_deprecated = true
			}
			`,
		},
	}

	for _, testCase := range testCases {
		expectError, _ := regexp.Compile(testCase.ExpectError)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      providerConfig + testCase.Config,
					ExpectError: expectError,
				},
			},
		})
	}
}
