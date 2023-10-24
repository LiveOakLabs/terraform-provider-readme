package readme

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/liveoaklabs/readme-api-go-client/readme"
	"github.com/liveoaklabs/terraform-provider-readme/internal/testdata"
)

func TestAPISpecificationsDataSource(t *testing.T) {
	testCases := []struct {
		name         string                    // The name of the test.
		config       string                    // The Terraform config to use for the test.
		expectError  string                    // If set, the test will expect an error matching this regex to be returned.
		response     []readme.APISpecification // List of specs that should be returned.
		responseCode int                       // Expected HTTP response code. Defaults to 200.
		excluded     []readme.APISpecification // List of specs that should explicitly *not* be returned.
	}{
		{
			name:     "it should return a list of all API specs when no filters are provided",
			config:   `data "readme_api_specifications" "test" {}`,
			response: testdata.APISpecifications,
		},

		// === Sorting.
		{
			name:     "it should return a list of unsorted API specs when sort_by is not provided",
			config:   `data "readme_api_specifications" "test" {}`,
			response: testdata.APISpecifications,
		},
		{
			name: "it should return a list of API specs sorted by title in ascending order when sort_by=title is provided",
			config: `data "readme_api_specifications" "test" {
				sort_by = "title"
			}`,
			response: []readme.APISpecification{
				testdata.APISpecifications[1],
				testdata.APISpecifications[0],
				testdata.APISpecifications[2],
			},
		},
		{
			name: "it should return a list of API specs sorted by last_synced in ascending order when sort_by=last_synced is provided",
			config: `data "readme_api_specifications" "test" {
				sort_by = "last_synced"
			}`,
			response: []readme.APISpecification{
				testdata.APISpecifications[2],
				testdata.APISpecifications[1],
				testdata.APISpecifications[0],
			},
		},

		// ==== Filter by spec title.
		{
			name: "it should return API specs that match the provided title filter",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = { title = ["%s", "%s"] }
			}`,
				testdata.APISpecifications[0].Title,
				testdata.APISpecifications[1].Title,
			),
			response: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[1],
			},
			excluded: []readme.APISpecification{testdata.APISpecifications[2]},
		},

		// ==== Filter by spec version ID.
		{
			name: "it should return API specs that match the provided version filter",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = { version = ["%s", "%s"] }
			}`,
				testdata.APISpecifications[0].Version,
				testdata.APISpecifications[1].Version,
			),
			response: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[1],
			},
			excluded: []readme.APISpecification{testdata.APISpecifications[2]},
		},

		// ==== Filter by has_category.
		{
			name: "it should return API specs that match the provided has_category filter",
			config: `data "readme_api_specifications" "test" {
				filter = { has_category = true }
			}`,
			response: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[1],
			},
			excluded: []readme.APISpecification{testdata.APISpecifications[2]},
		},
		{
			name: "it should return only API specs that have a category when has_category is true and used with another filter",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = {
					has_category  = true
					category_slug = ["%s", "%s", "test-api-spec-without-category"]
				}
			}`,
				testdata.APISpecifications[0].Category.Slug,
				testdata.APISpecifications[1].Category.Slug,
			),
			response: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[1],
			},
			excluded: []readme.APISpecification{testdata.APISpecifications[2]},
		},
		{
			name: "it should return an empty list when has_category is true and used with another filter that doesn't match any API specs",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category  = true
					category_slug = ["test-api-spec-without-category"]
				}
			}`,
			response: []readme.APISpecification{},
		},

		// ==== Filter by category title.
		{
			name: "it should return API specs that match the provided category_title filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["Test API Spec"] }
			}`,
			response: []readme.APISpecification{testdata.APISpecifications[0]},
			excluded: []readme.APISpecification{
				testdata.APISpecifications[1],
				testdata.APISpecifications[2],
			},
		},

		// ==== Filter by category ID.
		{
			name: "it should return API specs that match the provided category_id filter when has_category is not provided",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = { category_id = ["%s"] }
			}`,
				testdata.APISpecifications[1].Category.ID,
			),
			response: []readme.APISpecification{testdata.APISpecifications[1]},
			excluded: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[2],
			},
		},
		{
			name: "it should return API specs that match the provided category_id filter when has_category is true",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = {
					has_category = true
					category_id  = ["%s", "%s"]
				}
			}`,
				testdata.APISpecifications[0].Category.ID,
				testdata.APISpecificationsNoCategory.Category.ID,
			),
			response: []readme.APISpecification{testdata.APISpecifications[0]},
			excluded: []readme.APISpecification{
				testdata.APISpecifications[1],
				testdata.APISpecificationsNoCategory,
			},
		},

		// ==== Filter by category slug.
		{
			name: "it should return API specs that match the provided category_slug filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_slug = ["another-test-api-spec"] }
			}`,
			response: []readme.APISpecification{testdata.APISpecifications[1]},
			excluded: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[2],
			},
		},

		// ==== Filter combinations.
		{
			name: "it should return API specs that match the provided category_id and category_title filters",
			config: fmt.Sprintf(`data "readme_api_specifications" "test" {
				filter = {
					category_id    = ["%s"]
					category_title = ["%s"]
					has_category   = true
				}
			}`,
				testdata.APISpecifications[1].Category.ID,
				testdata.APISpecificationsNoCategory.Category.Title,
			),
			response: []readme.APISpecification{testdata.APISpecifications[1]},
			excluded: []readme.APISpecification{
				testdata.APISpecifications[0],
				testdata.APISpecifications[2],
			},
		},
		{
			name: "it does not return specs that don't match any filters provided",
			config: `data "readme_api_specifications" "test" {
				filter = {
					has_category   = true
					category_id    = ["000000011111122222223333"]
					category_slug  = ["this-doesnt-exist"]
					category_title = ["This doesn't exist"]
					title          = ["This doesn't exist either"]
				}
			}`,
			response: []readme.APISpecification{},
		},
		{
			name: "it does not return API specs that don't match the category_slug filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_slug = ["test-api-spec"] }
			}`,
			response: []readme.APISpecification{},
		},
		{
			name: "it does not return API specs that don't match the category_title filter",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["This does not exist"] }
			}`,
			response: []readme.APISpecification{},
		},

		// ==== Error handling.
		{
			name: "it should return an error when the API response is invalid",
			config: `data "readme_api_specifications" "test" {
				filter = { category_title = ["Test API Spec"] }
			}`,
			response:     []readme.APISpecification{},
			responseCode: 400,
			expectError:  "Unable to retrieve API specifications",
		},
	}

	for _, tc := range testCases { // nolint:varnamelen
		t.Run(tc.name, func(t *testing.T) {
			response := testdata.APISpecifications
			if tc.response != nil {
				response = tc.response
			}

			responseCode := 200
			if tc.responseCode != 0 {
				responseCode = tc.responseCode
			}

			if tc.expectError != "" {
				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: testProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config:      providerConfig + tc.config,
							ExpectError: regexp.MustCompile(tc.expectError),
							PreConfig:   testdata.APISpecificationRespond(response, responseCode),
						},
					},
				})
			} else {
				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: testProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: providerConfig + tc.config,
							Check: resource.ComposeAggregateTestCheckFunc(
								apiSpecificationsChecks(t, response, tc.excluded)...,
							),
							PreConfig: testdata.APISpecificationRespond(response, responseCode),
						},
					},
				})
			}
		})
	}
}

// apiSpecificationsChecks returns a list of TestCheckFuncs to test the data source.
// It ensures the number of specs returned is correct, the excluded specs are not returned,
// and the specs returned are correct.
func apiSpecificationsChecks(t *testing.T, specs, excluded []readme.APISpecification) []resource.TestCheckFunc {
	var checks []resource.TestCheckFunc

	// Ensure the number of specs returned is correct.
	checks = append(checks, resource.TestCheckResourceAttr(
		"data.readme_api_specifications.test",
		"specs.#",
		strconv.Itoa(len(specs)),
	))

	// Ensure the excluded specs are not returned.
	for _, exclude := range excluded {
		for _, spec := range specs {
			if spec.ID == exclude.ID {
				t.Fatalf("Expected to exclude API specification with ID: %s", spec.ID)
			}
		}
	}

	// Ensure the specs returned are correct.
	for i, spec := range specs { // nolint:varnamelen
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.id", i),
			spec.ID,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.last_synced", i),
			spec.LastSynced,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.source", i),
			spec.Source,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.title", i),
			spec.Title,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.type", i),
			spec.Type,
		))
		checks = append(checks, resource.TestCheckResourceAttr(
			"data.readme_api_specifications.test",
			fmt.Sprintf("specs.%d.version", i),
			spec.Version,
		))
	}

	return checks
}
